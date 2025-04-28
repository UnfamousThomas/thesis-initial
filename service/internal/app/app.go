package app

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"net/http"
)

type App struct {
	Mux           *http.ServeMux
	DynamicClient *dynamic.DynamicClient
	ClientSet     *kubernetes.Clientset
}

// CreateApp is where the app struct is created and related sub-variables are initialized
func CreateApp() *App {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal("Could not create config: ", err)
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatal("Could not create client: ", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal("Could not create clientset: ", err)
	}

	a := App{
		Mux:           http.NewServeMux(),
		DynamicClient: client,
		ClientSet:     clientset,
	}
	return &a
}
