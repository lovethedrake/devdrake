package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/jsonmessage"
)

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
