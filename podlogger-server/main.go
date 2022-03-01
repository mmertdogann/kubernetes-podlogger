package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	LoggerType "podlogger-server/types"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	podLogger  LoggerType.PodLogger
	namespace  string
	appName    string = "PodloggerServer"
	identifier string
	PORT       string
)

func main() {
	clientset := connectToK8s()

	// Read namespace and port from command line
	flag.StringVar(&namespace, "n", "default", "namespace")
	flag.StringVar(&PORT, "p", "8080", "port")
	flag.Parse()

	// Create a clientset interface specifically for pod access in the given namespace
	podInterface := clientset.CoreV1().Pods(namespace)

	// Create a watcher on pods
	podWatcher, err := podInterface.Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	// Create a pod event channel
	podChannel := podWatcher.ResultChan()

	identifier = fmt.Sprintf("******* From %s app and %s namespace *******", appName, namespace)

	fmt.Println("Server started, waiting for connection from client")

	log.Printf("Starting a %s in namespace %s\n\n", appName, namespace)

	log.Println(fmt.Sprintf("%s\n\n", identifier))
	echoLog()

	// Listen for connections from the client
	http.ListenAndServe(":"+PORT, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Client connected")
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			fmt.Println("Error starting socket server: " + err.Error())
		}

		// Read log from other namespace
		go func() {
			defer conn.Close()
			for {
				msg, _, err := wsutil.ReadClientData(conn)
				if err != nil {
					fmt.Println("Client disconnected")
					return
				}
				log.Println(string(msg))
			}
		}()

		// Watch loop
		for event := range podChannel {
			pod, ok := event.Object.(*v1.Pod)
			if !ok {
				log.Fatal(err)
			}

			var msg []byte

			switch event.Type {
			case watch.Added:
				log.Println(identifier)
				log.Printf("Pod added: %s \n", pod.Name)
				addPodLogger(pod)
				msg = []byte(fmt.Sprintf("From the %s namespace\n", namespace) + fmt.Sprintf("Pod added %s\n", pod.Name) + sendLog())
			case watch.Deleted:
				log.Println(identifier)
				log.Printf("Pod deleted: %s \n", pod.Name)
				deletePodLogger(pod.Name)
				msg = []byte(fmt.Sprintf("From the %s namespace\n", namespace) + fmt.Sprintf("Pod deleted %s\n", pod.Name) + sendLog())
			case watch.Modified:
				updatePodLogger(pod)
				msg = []byte(fmt.Sprintf("From the %s namespace\n", namespace) + fmt.Sprintf("Pod updated %s\n", pod.Name))
			}

			// Send log to the other namespace
			err = wsutil.WriteServerMessage(conn, ws.OpText, msg)
		}
	}))
}

func populateLog(podLoggerContainer *LoggerType.PodLoggerContainer, podLoggerTemplate *LoggerType.PodLoggerTemplate, pod *v1.Pod) {
	numOfContainers := len(pod.Spec.Containers)
	for _, container := range pod.Spec.Containers {
		podLoggerContainer.Name = container.Name
		podLoggerContainer.Image = container.Image
		podLoggerTemplate.PodLoggerContainer = append(podLoggerTemplate.PodLoggerContainer, *podLoggerContainer)
	}
	podLoggerTemplate.PodName = pod.Name
	podLoggerTemplate.ContainerSize = numOfContainers
}

func addPodLogger(pod *v1.Pod) {
	var podLoggerContainer LoggerType.PodLoggerContainer
	var podLoggerTemplate LoggerType.PodLoggerTemplate

	populateLog(&podLoggerContainer, &podLoggerTemplate, pod)

	podLogger.PodSize++
	podLogger.PodLoggerTemplate = append(podLogger.PodLoggerTemplate, podLoggerTemplate)

	echoLog()
}

// Pods are updated multiple times immediately after being created, so expect multiple calls for the same pod.
func updatePodLogger(pod *v1.Pod) {
	var podLoggerContainer LoggerType.PodLoggerContainer
	var podLoggerTemplate LoggerType.PodLoggerTemplate

	populateLog(&podLoggerContainer, &podLoggerTemplate, pod)

	for _, v := range podLogger.PodLoggerTemplate {
		if v.PodName == podLoggerTemplate.PodName {

			v.PodName = podLoggerTemplate.PodName
			v.ContainerSize = podLoggerTemplate.ContainerSize
			copy(v.PodLoggerContainer, podLoggerTemplate.PodLoggerContainer)
		}
	}

	log.Println(identifier)
	log.Printf("Pod updated: %s\n", podLoggerTemplate.PodName)
}

func deletePodLogger(podName string) {
	for idx, v := range podLogger.PodLoggerTemplate {
		if v.PodName == podName {
			podLogger.PodLoggerTemplate = append(podLogger.PodLoggerTemplate[:idx], podLogger.PodLoggerTemplate[idx+1:]...)
			podLogger.PodSize--
		}
	}
	echoLog()
}

func sendLog() string {
	var msg string
	out, err := json.MarshalIndent(podLogger, "", "  ")
	if err != nil {
		panic(err)
	}
	if string(out) != "{}" {
		msg = fmt.Sprintf("From the %s namespace\nPods' Status:\n%s\n", namespace, string(out))
		return msg
	} else {
		msg = fmt.Sprintf("There is no pod in the current namespace %s\n\n", namespace)
		return msg
	}

}

func echoLog() {
	out, err := json.MarshalIndent(podLogger, "", "  ")
	if err != nil {
		panic(err)
	}
	if string(out) != "{}" {
		log.Println(identifier)
		log.Printf("Pods' Status:\n%s\n", string(out))
	} else {
		log.Printf("There is no pod in the current namespace %s\n\n", namespace)
	}
}

func connectToK8s() *kubernetes.Clientset {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return clientset
}
