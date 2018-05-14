package kubernetes

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetOwnerReference(obj metav1.Object, owner metav1.Object) *metav1.OwnerReference {
	for _, ref := range obj.GetOwnerReferences() {
		if ref.UID == owner.GetUID() {
			return &ref
		}
	}
	return nil
}
