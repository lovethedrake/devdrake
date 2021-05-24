package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	dockerTypes "github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/lovethedrake/go-drake/config"
)

func (e *executor) shouldPull(
	ctx context.Context,
	imageName string,
	imagePullPolicy config.ImagePullPolicy,
) (bool, error) {
	if _, _, err :=
		e.dockerClient.ImageInspectWithRaw(ctx, imageName); err != nil {
		if docker.IsErrNotFound(err) {
			return true, nil
		}
		return false, err
	}
	return imagePullPolicy == config.ImagePullPolicyAlways, nil
}

func (e *executor) pullImages(
	ctx context.Context,
	imageNames map[string]struct{},
) error {
	for imageName := range imageNames {
		fmt.Printf("~~~~> pulling image %q <~~~~\n", imageName)
		reader, err := e.dockerClient.ImagePull(
			ctx,
			imageName,
			dockerTypes.ImagePullOptions{},
		)
		if err != nil {
			return err
		}
		defer reader.Close()
		dec := json.NewDecoder(reader)
		for {
			var message jsonmessage.JSONMessage
			if err := dec.Decode(&message); err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
			fmt.Println(message.Status)
		}
	}
	return nil
}
