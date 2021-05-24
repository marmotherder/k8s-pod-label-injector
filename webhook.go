package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	admission_v1 "k8s.io/api/admission/v1"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//patchReq is a struct for a patch request that doesn't exist in types from admission
type patchReq struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// hook is the main function for inbound hook requests
func hook(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "invalid Content-Type, expect `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	as := admission_v1.AdmissionResponse{}
	ar, pod, err := readRequest(r.Body)

	if err != nil {
		setResultMessage(&as, err.Error())
	}

	if ar != nil && ar.Request.Operation == admission_v1.Create && err == nil {
		if err != nil {
			setResultMessage(&as, err.Error())
		} else {
			err := runCreateHook(w, r, ar, pod, &as)
			if err != nil {
				setResultMessage(&as, err.Error())
			} else {
				as.Allowed = true
			}
		}
	}

	dispatchResponse(ar, as, w)
}

// addPatchReq adds a patchReq struct to a slice in a repeatable generic way
func addPatchReq(isNew func() bool, op string, path string, patchReqs *[]patchReq, value interface{}) {
	req := patchReq{
		Op:    op,
		Value: value,
	}

	if isNew() {
		req.Path = path
	} else {
		req.Path = fmt.Sprintf("%s/-", path)
	}
	*patchReqs = append(*patchReqs, req)
}

// runCreateHook is a split out function for running the steps when creating a pod
func runCreateHook(w http.ResponseWriter, r *http.Request, ar *admission_v1.AdmissionReview, pod *core_v1.Pod, as *admission_v1.AdmissionResponse) error {
	mutate, err := shouldMutate(pod)
	if err != nil {
		return err
	}
	if !mutate {
		return nil
	}

	_, ok := pod.Labels["fargate"]

	labels := map[string]string{"fargate": "enabled"}
	patchReqs := make([]patchReq, 0)
	addPatchReq(func() bool { return ok }, "add", "/metadata/labels", &patchReqs, labels)

	patchData, err := json.Marshal(patchReqs)
	if err != nil {
		return err
	}

	as.Patch = patchData
	as.PatchType = func() *admission_v1.PatchType {
		p := admission_v1.PatchTypeJSONPatch
		return &p
	}()

	setResultMessage(as, "Added label fargate=enabled to pod")
	return nil
}

// readRequest reads the inbound hook request
func readRequest(body io.ReadCloser) (*admission_v1.AdmissionReview, *core_v1.Pod, error) {
	var ar admission_v1.AdmissionReview
	if err := json.NewDecoder(body).Decode(&ar); err != nil {
		return nil, nil, err
	}
	if len(ar.Request.Object.Raw) > 0 {
		var pod core_v1.Pod
		if err := json.Unmarshal(ar.Request.Object.Raw, &pod); err != nil {
			return nil, nil, err
		}
		return &ar, &pod, nil
	}
	return &ar, nil, nil
}

// shouldMutate will scan the pod for the desired annotations
// If the annotations exist, return the paths we need to progress with the secrets creation
func shouldMutate(pod *core_v1.Pod) (bool, error) {
	for _, ref := range pod.OwnerReferences {
		if ref.Kind == "ReplicaSet" {
			client, err := NewK8SClient()
			if err != nil {
				return false, err
			}
			rs, err := client.AppsV1().ReplicaSets(pod.Namespace).Get(context.Background(), ref.Name, meta_v1.GetOptions{})
			if err != nil {
				return false, err
			}
			if *rs.Spec.Replicas == int32(0) {
				return true, nil
			}
		}
	}
	return false, nil
}

// setResultMessage adds a message to the response struct, can be errors or otherwise
func setResultMessage(as *admission_v1.AdmissionResponse, message string) {
	status := meta_v1.Status{Message: message}
	as.Result = &status
}

// dispatchResponse writes out the json response payload from the hook
func dispatchResponse(ar *admission_v1.AdmissionReview, as admission_v1.AdmissionResponse, w http.ResponseWriter) {
	resp := admission_v1.AdmissionReview{}
	resp.Response = &as
	if ar.Request != nil {
		resp.Response.UID = ar.Request.UID
	}
	payloadjson, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "could not decode request message", http.StatusInternalServerError)
	}
	log.Println(string(payloadjson))
	w.Write(payloadjson)
}
