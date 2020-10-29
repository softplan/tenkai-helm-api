package handlers

import (
	"encoding/base64"
	"log"
	"os"
	"strings"
)


func createEnvironmentFile(clusterName string, clusterUserToken string,
	fileName string, ca string, server string, namespace string) {

	removeEnvironmentFile(fileName)
	
	file, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	ca = strings.TrimSuffix(ca, "\n")
	caBase64 := base64.StdEncoding.EncodeToString([]byte(ca))

	startIndex := strings.Index(clusterUserToken, "kubeconfig-")
	clusterUser := "xpto"
	if startIndex > 0 {
		startIndex = startIndex + 11
		endIndex := strings.Index(clusterUserToken, ":")
		clusterUser = clusterUserToken[startIndex:endIndex]
	}

	file.WriteString("apiVersion: v1\n")
	file.WriteString("clusters:\n")
	file.WriteString("- cluster:\n")
	file.WriteString("    certificate-authority-data: " + caBase64 + "\n")
	file.WriteString("    server: " + server + "\n")
	file.WriteString("  name: " + clusterName + "\n")
	file.WriteString("contexts:\n")
	file.WriteString("- context:\n")
	file.WriteString("    cluster: " + clusterName + "\n")
	file.WriteString("    namespace: " + namespace + "\n")
	file.WriteString("    user: " + clusterUser + "\n")
	file.WriteString("  name: " + clusterName + "\n")
	file.WriteString("current-context: " + clusterName + "\n")
	file.WriteString("kind: Config\n")
	file.WriteString("preferences: {}\n")
	file.WriteString("users:\n")
	file.WriteString("- name: " + clusterUser + "\n")
	file.WriteString("  user:\n")
	file.WriteString("    token: " + clusterUserToken + "\n")

}

func removeEnvironmentFile(fileName string) error {
	log.Println("Removing file: " + fileName)

	if _, err := os.Stat("./" + fileName); err == nil {
		err := os.Remove("./" + fileName)
		if err != nil {
			log.Println("Error removing file", err)
			return err
		}
	}
	return nil
}