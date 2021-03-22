package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/ironcore864/mydumpergo/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/kelseyhightower/envconfig"
)

func check(cmd *exec.Cmd, outputdir, region, bucket string, maxFileCount int) {
	// s3 manager
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	uploader := s3manager.NewUploader(sess)

	for {
		files, _ := ioutil.ReadDir(outputdir)
		if len(files) > maxFileCount {
			// pause mydumper process when there are more files than configured threshold
			cmd.Process.Signal(syscall.SIGTSTP)

			// upload each file to s3
			for i := 0; i < len(files); i++ {
				filename := files[i].Name()
				file, err := os.Open(outputdir + "/" + filename)
				if err != nil {
					log.Printf("Unable to open file %s, %s\n", filename, err.Error())
				}
				defer file.Close()

				_, err = uploader.Upload(&s3manager.UploadInput{
					Bucket: aws.String(bucket),
					Key:    aws.String(filename),
					Body:   file,
				})
				if err != nil {
					log.Printf("Unable to upload %s to %s, %s\n", filename, bucket, err.Error())
				}
				log.Printf("Successfully uploaded %s to %s\n", filename, bucket)
			}

			// delete uploaded files
			for i := 0; i < len(files); i++ {
				filename := files[i].Name()
				os.Remove(outputdir + "/" + filename)
			}

			// continue mydumper process
			cmd.Process.Signal(syscall.SIGCONT)
		}
	}
}

func cleanup(outputDir string) {
	files, _ := ioutil.ReadDir(outputDir)
	for i := 0; i < len(files); i++ {
		filename := files[i].Name()
		os.Remove(outputDir + "/" + filename)
	}
	os.Remove(outputDir)
}

func main() {
	// load config
	var c config.Conf
	err := envconfig.Process("", &c)
	if err != nil {
		log.Fatal(err.Error())
	}

	// mkdir if doesn't exist
	if _, err := os.Stat(c.OutputDir); os.IsNotExist(err) {
		os.Mkdir(c.OutputDir, os.ModeDir|0755)
	}

	// start command
	cmd := exec.Command("mydumper",
		"-h", c.Host,
		"-P", c.Port,
		"-u", c.Username,
		"-p", c.Password,
		"-o", c.OutputDir,
		"-F", c.ChunkFileSizeMB)
	err = cmd.Start()
	if err != nil {
		log.Fatal(err.Error())
	}

	// start a goroutine which checks disk usage
	// if exceeds a threshold, pause the mydumper process, upload file to s3, delete files, then continue
	go check(cmd, c.OutputDir, c.Region, c.Bucket, c.MaxFileCount)

	// wait for command to finish
	err = cmd.Wait()
	if err != nil {
		log.Fatalf("Command finished with error: %s", err.Error())
	}

	// clean up remaining output files
	cleanup(c.OutputDir)
}
