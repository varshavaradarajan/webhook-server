package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/util/json"
	"log"
	"net/http"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	fmt.Println("starting webhook server")
	http.HandleFunc("/update-image", updateImage)
	http.ListenAndServe(":8000", nil)

}

func updateImage(writer http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	log.Println("reading request body")
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("unmarshaling request payload to map")
	data := map[string]interface{}{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Fatalln(err)
	}

	registry := data["registry_package"]
	packageVersion := registry.(map[string]interface{})["package_version"]
	url := packageVersion.(map[string]interface{})["package_url"]

	image := fmt.Sprintf("%s", url)
	log.Printf("updating pods to use image: %s\n", image)

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalln(err)
	}

	clientset , err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalln(err)
	} else {
		log.Println("connect k8s success")
	}

	pods, err := clientset.CoreV1().Pods("default").List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatalln(err, "unable to list pods")
	}

	for _, pod := range pods.Items {
		pod := pod
		for _, container := range pod.Spec.Containers {
			if strings.Split(container.Image, ":")[0] == strings.Split(image, ":")[0] {
				log.Printf("container image changed for pod: %s\n", pod.Name)
				container.Image = image
			}
		}
		log.Printf("updating pod %s\n", pod.Name)

		updatedPod, err := clientset.CoreV1().Pods("default").Update(ctx, &pod, metav1.UpdateOptions{})
		if err != nil {
			log.Fatalln(err)
		}

		log.Println(updatedPod)
	}
}
