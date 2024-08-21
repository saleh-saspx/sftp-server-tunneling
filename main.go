package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"golang.org/x/crypto/ssh"
)

// ساختار JSON برای دریافت اطلاعات سرور A و B
type ServerInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Port     string `json:"port"`
}

type RequestPayload struct {
	AServer ServerInfo `json:"a_server"`
	BServer ServerInfo `json:"b_server"`
}

func main() {
	http.HandleFunc("/copyfile", func(w http.ResponseWriter, r *http.Request) {
		// پردازش درخواست HTTP
		var payload RequestPayload
		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		jumpHost := fmt.Sprintf("%s:%s", "a_Ip", payload.AServer.Port)
		jumpUser := payload.AServer.Username
		jumpPassword := payload.AServer.Password

		targetHost := fmt.Sprintf("%s:%s", "b_Ip", payload.BServer.Port)
		targetUser := payload.BServer.Username
		targetPassword := payload.BServer.Password

		remoteFilePath := "/media/mmcblk0p1/Ringbuffers/yourfile"
		localFilePath := "/tmp/yourfile"

		// اتصال به سرور A
		jumpClient, err := connectToJumpHost(jumpHost, jumpUser, jumpPassword)
		if err != nil {
			log.Fatalf("Failed to connect to jump host: %v", err)
		}
		defer jumpClient.Close()

		// اتصال به سرور B از طریق سرور A
		targetClient, err := connectToTargetHost(jumpClient, targetHost, targetUser, targetPassword)
		if err != nil {
			log.Fatalf("Failed to connect to target host: %v", err)
		}
		defer targetClient.Close()

		// کپی کردن فایل از سرور B به سرور A
		err = copyFileFromRemote(targetClient, remoteFilePath, localFilePath)
		if err != nil {
			log.Fatalf("Failed to copy file: %v", err)
		}

		fmt.Fprintf(w, "File copied successfully to %s", localFilePath)
	})

	fmt.Println("Starting server on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// اتصال به سرور A
func connectToJumpHost(jumpHost, jumpUser, jumpPassword string) (*ssh.Client, error) {
	jumpConfig := &ssh.ClientConfig{
		User: jumpUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(jumpPassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	jumpClient, err := ssh.Dial("tcp", jumpHost, jumpConfig)
	if err != nil {
		return nil, err
	}

	return jumpClient, nil
}

// اتصال به سرور B از طریق سرور A
func connectToTargetHost(jumpClient *ssh.Client, targetHost, targetUser, targetPassword string) (*ssh.Client, error) {
	conn, err := jumpClient.Dial("tcp", targetHost)
	if err != nil {
		return nil, err
	}

	targetConfig := &ssh.ClientConfig{
		User: targetUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(targetPassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	ncc, chans, reqs, err := ssh.NewClientConn(conn, targetHost, targetConfig)
	if err != nil {
		return nil, err
	}

	targetClient := ssh.NewClient(ncc, chans, reqs)
	return targetClient, nil
}

// کپی کردن فایل از سرور B به سرور A
func copyFileFromRemote(client *ssh.Client, remoteFilePath, localFilePath string) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	// ایجاد فایل محلی برای ذخیره
	outputFile, err := os.Create(localFilePath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	// اجرای دستور برای خواندن فایل از راه دور
	output, err := session.StdoutPipe()
	if err != nil {
		return err
	}

	if err := session.Start(fmt.Sprintf("cat %s", remoteFilePath)); err != nil {
		return err
	}

	// نوشتن فایل به صورت محلی
	if _, err := io.Copy(outputFile, output); err != nil {
		return err
	}

	return session.Wait()
}
