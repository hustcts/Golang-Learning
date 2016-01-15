package main

import (
	"os"

	"github.com/ChengTiesheng/oci2docker/convert"
	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

func main() {
	logrus.SetLevel(logrus.InfoLevel)
	app := cli.NewApp()
	app.Name = "oci2docker"
	app.Usage = "A tool for coverting oci bundle to docker image"
	app.Version = "0.1.0"
	app.Commands = []cli.Command{
		{
			Name:  "convert",
			Usage: "convert operation",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "oci-bundle",
					Value: ".",
					Usage: "path of oci-bundle to convert",
				},
			},
			Action: oci2docker,
		},
	}

	app.Run(os.Args)
}

func oci2docker(c *cli.Context) {
	ociPath := c.String("oci-bundle")

	_, err := os.Stat(ociPath)
	if os.IsNotExist(err) {
		cli.ShowCommandHelp(c, "convert")
	}

	convert.RunOCI2Docker(ociPath)
}




package convert

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

type DockerInfo struct {
	Appdir     string
	Entrypoint string
	Expose     string
}

const (
	buildTemplate = `
FROM scratch
MAINTAINER ChengTiesheng <chengtiesheng@huawei.com>
ENTRYPOINT ["{{.Entrypoint}}"]
ADD {{.Appdir}} .
EXPOSE {{.Expose}}
`
)

func RunOCI2Docker(path string) error {
	appdir := path + "/rootfs"
	entrypoint := getEntrypointFromSpecs(path)

	dockerInfo := DockerInfo{
		Appdir:     appdir,
		Entrypoint: entrypoint,
		Expose:     "",
	}

	generateDockerfile(dockerInfo)

	dirWork := createWorkDir()

	exec.Command("mv", "./Dockerfile", dirWork)
	exec.Command("cp", "-rf", path+"/rootfs", dirWork)

	return nil
}

func generateDockerfile(dockerInfo DockerInfo) {
	t := template.Must(template.New("buildTemplate").Parse(buildTemplate))

	f, err := os.Create("Dockerfile")
	if err != nil {
		log.Fatal("Error wrinting Dockerfile %v", err.Error())
		return
	}
	defer f.Close()

	t.Execute(f, dockerInfo)

	fmt.Printf("Dockerfile generated, you can build the image with: \n")
	fmt.Printf("$ docker build -t %s .\n", dockerInfo.Entrypoint)

	return
}

// Create work directory for the conversion output
func createWorkDir() string {
	idir, err := ioutil.TempDir("", "oci2docker")
	if err != nil {
		return ""
	}
	rootfs := filepath.Join(idir, "rootfs")
	os.MkdirAll(rootfs, 0755)

	data := []byte{}
	if err := ioutil.WriteFile(filepath.Join(idir, "Dockerfile"), data, 0644); err != nil {
		return ""
	}
	return idir
}

func getEntrypointFromSpecs(path string) string {
	return "/bin/sh"
}

