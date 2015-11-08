package main

import (
	"github.com/fsouza/go-dockerclient"

	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type Image struct {
	Id          string
	ParentId    string   `json:",omitempty"`
	RepoTags    []string `json:",omitempty"`
	VirtualSize int64
	Size        int64
	Created     int64
}

type ImagesCommand struct {
	Dot          bool `short:"d" long:"dot" description:"Show image information as Graphviz dot. You can add a start image id or name -d/--dot [id/name]"`
	Tree         bool `short:"t" long:"tree" description:"Show image information as tree. You can add a start image id or name -t/--tree [id/name]"`
	Short        bool `short:"s" long:"short" description:"Show short summary of images (repo name and list of tags)."`
	NoTruncate   bool `short:"n" long:"no-trunc" description:"Don't truncate the image IDs."`
	OnlyLabelled bool `short:"l" long:"only-labelled" description:"Print only labelled images/containers."`
}

var imagesCommand ImagesCommand


func (x *ImagesCommand) Execute(args []string) error {
	var images *[]Image

	stat, err := os.Stdin.Stat()
	if err != nil {
		return fmt.Errorf("error reading stdin stat", err)
	}

	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// read in stdin
		stdin, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("error reading all input", err)
		}

		images, err = parseImagesJSON(stdin)
		if err != nil {
			return err
		}

	} else {

		client, err := connect()
		if err != nil {
			return err
		}

		clientImages, err := client.ListImages(docker.ListImagesOptions{All: true})
		if err != nil {
			if in_docker := os.Getenv("IN_DOCKER"); len(in_docker) > 0 {
				return fmt.Errorf("Unable to access Docker socket, please run like this:\n  docker run --rm -v /var/run/docker.sock:/var/run/docker.sock nate/dockviz images <args>\nFor more help, run 'dockviz help'")
			} else {
				return fmt.Errorf("Unable to connect: %s\nFor help, run 'dockviz help'", err)
			}
		}

		var ims []Image
		for _, image := range clientImages {
			// fmt.Println(image)
			ims = append(ims, Image{
				image.ID,
				image.ParentID,
				image.RepoTags,
				image.VirtualSize,
				image.Size,
				image.Created,
			})
		}

		images = &ims
	}

	 if imagesCommand.Tree || imagesCommand.Dot {
		var startImage *Image
		if len(args) > 0 {

			// attempt to find the start image, which can be specified as an
			// image ID or a repository name
			startImageArg := args[0]
			startImageRepo := args[0]

			// if tag is not defined, find by :latest tag
			if strings.Index(startImageRepo, ":") == -1 {
				startImageRepo = fmt.Sprintf("%s:latest", startImageRepo)
			}

		IMAGES:
			for _, image := range *images {
				// find by image id
				if strings.Index(image.Id, startImageArg) == 0 {
					startImage = &image
					break IMAGES
				}

				// find by image name (name and tag)
				for _, repotag := range image.RepoTags {
					if repotag == startImageRepo {
						startImage = &image
						break IMAGES
					}
				}
			}

			if startImage == nil {
				return fmt.Errorf("Unable to find image %s = %s.", startImageArg, startImageRepo)
			}
		}
		
		// select the start image of the tree
		var roots []Image
		if startImage == nil {
			roots = collectRoots(images)
		} else {
			startImage.ParentId = ""
			roots = []Image{*startImage}
		}
		
		// build helper map (image -> children)
		var imagesByParent = make(map[string][]Image)
		imagesByParent = collectChildren(images);
		
		// image ids truncate
		// initialize image informations
		
		// filter images
		if imagesCommand.OnlyLabelled{
			*images, imagesByParent = filterImages(images, &imagesByParent)
		}
		
		var buffer bytes.Buffer
		
		if imagesCommand.Tree {
			jsonToText(&buffer, imagesCommand.NoTruncate, roots, imagesByParent, "") 
		}
		if imagesCommand.Dot {
			buffer.WriteString("digraph docker {\n")
			imagesToDot(&buffer, roots, imagesByParent)
			buffer.WriteString(" base [style=invisible]\n}\n")
			buffer.String()
		}
		
		fmt.Print(buffer.String())
	} else if imagesCommand.Short {
		fmt.Printf(jsonToShort(images))
	} else {
		return fmt.Errorf("Please specify either --dot, --tree, or --short")
	}

	return nil
}


func collectChildren(images *[]Image) map[string][]Image {
	var imagesByParent = make(map[string][]Image)
	for _, image := range *images {
		if children, exists := imagesByParent[image.ParentId]; exists {
			imagesByParent[image.ParentId] = append(children, image)
		} else {
			imagesByParent[image.ParentId] = []Image{image}
		}
	}
	
	return imagesByParent
}


func collectRoots(images *[]Image) []Image {
	var roots []Image
	for _, image := range *images {
		if image.ParentId == "" {
			roots = append(roots, image)
		}
	}
	
	return roots
}


func filterImages (images *[]Image, byParent *map[string][]Image) (filteredImages []Image, filteredChildren map[string][]Image) {
	for i := 0; i<len(*images); i++ {
		// image is visible 
		//   1. it has a label
		//   2. it is root
		//   3. it is a node 
		var visible bool = (*images)[i].RepoTags[0] != "<none>:<none>" || (*images)[i].ParentId == "" || len((*byParent)[(*images)[i].Id]) > 1
		if visible {
			filteredImages = append(filteredImages, (*images)[i])
		} else {
			// change childs parent id
			// if items are filtered with only one child
			for j := 0; j<len(filteredImages); j++ {
				if filteredImages[j].ParentId == (*images)[i].Id {
					filteredImages[j].ParentId = (*images)[i].ParentId
				}
			}
			for j := 0; j<len(*images); j++ {
				if (*images)[j].ParentId == (*images)[i].Id {
					(*images)[j].ParentId = (*images)[i].ParentId
				}
			}
		}
	}

	filteredChildren = collectChildren(&filteredImages)
	
	return filteredImages, filteredChildren
}

func jsonToText(buffer *bytes.Buffer, noTrunc bool, images []Image, byParent map[string][]Image, prefix string) {
	var length = len(images)
	if length > 1 {
		for index, image := range images {
			var nextPrefix string = ""
			if index+1 == length {
				PrintTreeNode(buffer, noTrunc, image, prefix+"└─")
				nextPrefix = "  "
			} else {
				PrintTreeNode(buffer, noTrunc, image, prefix+"├─")
				nextPrefix = "│ "
			}
			if subimages, exists := byParent[image.Id]; exists {
				jsonToText(buffer, noTrunc, subimages, byParent, prefix+nextPrefix)
			}
		}
	} else {
		for _, image := range images {
			PrintTreeNode(buffer, noTrunc, image, prefix+"└─")
			if subimages, exists := byParent[image.Id]; exists {
				jsonToText(buffer, noTrunc, subimages, byParent, prefix+"  ")
			}
		}
	}
}


func PrintTreeNode(buffer *bytes.Buffer, noTrunc bool, image Image, prefix string) {
	var imageID string
	if noTrunc {
		imageID = image.Id
	} else {
		imageID = truncate(image.Id)
	}

	buffer.WriteString(fmt.Sprintf("%s%s Virtual Size: %s", prefix, imageID, humanSize(image.VirtualSize)))
	if image.RepoTags[0] != "<none>:<none>" {
		buffer.WriteString(fmt.Sprintf(" Tags: %s\n", strings.Join(image.RepoTags, ", ")))
	} else {
		buffer.WriteString(fmt.Sprintf("\n"))
	}
}

func humanSize(raw int64) string {
	sizes := []string{"B", "KB", "MB", "GB", "TB"}

	rawFloat := float64(raw)
	ind := 0

	for {
		if rawFloat < 1000 {
			break
		} else {
			rawFloat = rawFloat / 1000
			ind = ind + 1
		}
	}

	return fmt.Sprintf("%.01f %s", rawFloat, sizes[ind])
}


func truncate(id string) string {
	return id[0:12]
}


func parseImagesJSON(rawJSON []byte) (*[]Image, error) {

	var images []Image
	err := json.Unmarshal(rawJSON, &images)

	if err != nil {
		return nil, fmt.Errorf("Error reading JSON: ", err)
	}

	return &images, nil
}


func imagesToDot(buffer *bytes.Buffer, images []Image, byParent map[string][]Image) {
	for _, image := range images {
		if image.ParentId == "" {
			buffer.WriteString(fmt.Sprintf(" base -> \"%s\" [style=invis]\n", truncate(image.Id)))
		} else {
			buffer.WriteString(fmt.Sprintf(" \"%s\" -> \"%s\"\n", truncate(image.ParentId), truncate(image.Id)))
		}
		if image.RepoTags[0] != "<none>:<none>" {
			buffer.WriteString(fmt.Sprintf(" \"%s\" [label=\"%s\\n%s\",shape=box,fillcolor=\"paleturquoise\",style=\"filled,rounded\"];\n", truncate(image.Id), truncate(image.Id), strings.Join(image.RepoTags, "\\n")))
		}
		if subimages, exists := byParent[image.Id]; exists {
			imagesToDot(buffer, subimages, byParent)
		}
	}
}


func jsonToShort(images *[]Image) string {
	var buffer bytes.Buffer

	var byRepo = make(map[string][]string)

	for _, image := range *images {
		for _, repotag := range image.RepoTags {
			if repotag != "<none>:<none>" {

				// parse the repo name and tag name out
				// tag is after the last colon
				lastColonIndex := strings.LastIndex(repotag, ":")
				tagname := repotag[lastColonIndex+1:]
				reponame := repotag[0:lastColonIndex]

				if tags, exists := byRepo[reponame]; exists {
					byRepo[reponame] = append(tags, tagname)
				} else {
					byRepo[reponame] = []string{tagname}
				}
			}
		}
	}

	for repo, tags := range byRepo {
		buffer.WriteString(fmt.Sprintf("%s: %s\n", repo, strings.Join(tags, ", ")))
	}

	return buffer.String()
}


func init() {
	parser.AddCommand("images",
		"Visualize docker images.",
		"",
		&imagesCommand)
}
