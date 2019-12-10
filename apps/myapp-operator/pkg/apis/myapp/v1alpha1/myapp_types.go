package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MyAppSpec struct {
	DatabaseType  string `json:"database_type"`
	DatabaseName  string `json:"database_name"`
	Host          string `json:"host"`
	Port          int32  `json:"port"`
	User          string `json:"user"`
	Pass          string `json:"pass"`
	ContainerPort int32  `json:"container_port"`
}

type MyAppStatus struct {
}

type MyApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MyAppSpec   `json:"spec,omitempty"`
	Status MyAppStatus `json:"status,omitempty"`
}

// MyAppList contains a list of MyApp
type MyAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MyApp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MyApp{}, &MyAppList{})
}
