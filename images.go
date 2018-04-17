package main

import (
	"github.com/dustin/go-humanize"
	"github.com/fsouza/go-dockerclient"

	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type Image struct {
	Id          string
	ParentId    string   `json:",omitempty"`
	RepoTags    []string `json:",omitempty"`
	VirtualSize int64
	Size        int64
	Created     int64
	OrigId      string
	CreatedBy   string
}

type ImagesCommand struct {
	Dot          bool `short:"d" long:"dot" description:"Show image information as Graphviz dot. You can add a start image id or name -d/--dot [id/name]"`
	Tree         bool `short:"t" long:"tree" description:"Show image information as tree. You can add a start image id or name -t/--tree [id/name]"`
	Short        bool `short:"s" long:"short" description:"Show short summary of images (repo name and list of tags)."`
	NoTruncate   bool `short:"n" long:"no-trunc" description:"Don't truncate the image IDs (only works with tree mode)."`
	Incremental  bool `short:"i" long:"incremental" description:"Display image size as incremental rather than cumulative."`
	OnlyLabelled bool `short:"l" long:"only-labelled" description:"Print only labelled images/containers."`
	NoHuman      bool `short:"c" long:"no-human" description:"Don't humanize the sizes."`
}

type DisplayOpts struct {
	NoTruncate  bool
	Incremental bool
	NoHuman     bool
}

var imagesCommand ImagesCommand

func (x *ImagesCommand) Execute(args []string) error {
	var images *[]Image

	stat, err := os.Stdin.Stat()
	if err != nil {
		return fmt.Errorf("error reading stdin stat", err)
	}

	if globalOptions.Stdin && (stat.Mode()&os.ModeCharDevice) == 0 {
		// read in stdin
		stdin, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("error reading all input", err)
		}

		images, err = parseImagesJSON(stdin)
		if err != nil {
			return err
		}

		var ims []Image
		for _, image := range *images {
			ims = append(ims, Image{
				image.Id,
				image.ParentId,
				image.RepoTags,
				image.VirtualSize,
				image.Size,
				image.Created,
				image.Id,
				"",
			})
		}

		images = &ims

	} else {

		client, err := connect()
		if err != nil {
			return err
		}

		ver, err := getAPIVersion(client)
		if err != nil {
			if in_docker := os.Getenv("IN_DOCKER"); len(in_docker) > 0 {
				return fmt.Errorf("Unable to access Docker socket, please run like this:\n  docker run -it --rm -v /var/run/docker.sock:/var/run/docker.sock nate/dockviz images <args>\nFor more help, run 'dockviz help'")
			} else {
				return fmt.Errorf("Unable to connect: %s\nFor help, run 'dockviz help'", err)
			}
		}

		if ver[0] == 1 && ver[1] <= 21 {
			clientImages, err := client.ListImages(docker.ListImagesOptions{All: true})
			if err != nil {
				return err
			}

			var ims []Image
			for _, image := range clientImages {
				ims = append(ims, Image{
					image.ID,
					image.ParentID,
					image.RepoTags,
					image.VirtualSize,
					image.Size,
					image.Created,
					image.ID,
					"",
				})
			}

			images = &ims
		} else {
			clientImages, err := client.ListImages(docker.ListImagesOptions{})
			if err != nil {
				return err
			}

			images, err = synthesizeImagesFromHistory(client, clientImages)
			if err != nil {
				return err
			}
		}
	}

	if imagesCommand.Tree || imagesCommand.Dot {
		var startImage *Image
		if len(args) > 0 {
			startImage, err = findStartImage(args[0], images)

			if err != nil {
				return err
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
		imagesByParent := collectChildren(images)

		// filter images
		if imagesCommand.OnlyLabelled {
			*images, imagesByParent = filterImages(images, &imagesByParent)
		}

		if imagesCommand.Tree {
			dispOpts := DisplayOpts{
				imagesCommand.NoTruncate,
				imagesCommand.Incremental,
				imagesCommand.NoHuman,
			}
			fmt.Print(jsonToTree(roots, imagesByParent, dispOpts))
		}
		if imagesCommand.Dot {
			fmt.Print(jsonToDot(roots, imagesByParent))
		}

	} else if imagesCommand.Short {
		fmt.Printf(jsonToShort(images))
	} else {
		return fmt.Errorf("Please specify either --dot, --tree, or --short")
	}

	return nil
}

func synthesizeImagesFromHistory(client *docker.Client, images []docker.APIImages) (*[]Image, error) {
	var newImages []Image
	newImageRoster := make(map[string]*Image)
	for _, image := range images {
		var previous string
		var vSize int64
		history, err := client.ImageHistory(image.ID)
		if err != nil {
			return &newImages, err
		}
		for i := len(history) - 1; i >= 0; i-- {
			var newID string
			h := sha256.New()
			h.Write([]byte(previous))
			h.Write([]byte(history[i].CreatedBy))
			h.Write([]byte(strconv.FormatInt(history[i].Created, 10)))
			h.Write([]byte(strconv.FormatInt(history[i].Size, 10)))
			newID = fmt.Sprintf("synth:%s", hex.EncodeToString(h.Sum(nil)))

			vSize = vSize + history[i].Size
			existingImage, ok := newImageRoster[newID]
			if !ok {
				newImageRoster[newID] = &Image{
					newID,
					previous,
					history[i].Tags,
					vSize,
					history[i].Size,
					history[i].Created,
					history[i].ID,
					history[i].CreatedBy,
				}
			} else {
				if len(history[i].Tags) > 0 {
					existingImage.RepoTags = append(existingImage.RepoTags, history[i].Tags...)
				}
			}
			previous = newID
		}
	}

	for _, image := range newImageRoster {
		if len(image.RepoTags) == 0 {
			image.RepoTags = []string{"<none>:<none>"}
		} else {
			visited := make(map[string]bool)
			for _, tag := range image.RepoTags {
				visited[tag] = true
			}
			image.RepoTags = []string{}
			for tag, _ := range visited {
				image.RepoTags = append(image.RepoTags, tag)
			}
		}
		newImages = append(newImages, *image)
	}
	return &newImages, nil
}

func findStartImage(name string, images *[]Image) (*Image, error) {

	var startImage *Image

	// attempt to find the start image, which can be specified as an
	// image ID or a repository name
	startImageArg := name
	startImageRepo := name

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
		return nil, fmt.Errorf("Unable to find image %s = %s.", startImageArg, startImageRepo)
	}

	return startImage, nil
}

func jsonToTree(images []Image, byParent map[string][]Image, dispOpts DisplayOpts) string {
	var buffer bytes.Buffer

	jsonToText(&buffer, images, byParent, dispOpts, "")

	return buffer.String()
}

func jsonToDot(roots []Image, byParent map[string][]Image) string {
	var buffer bytes.Buffer

	buffer.WriteString("digraph docker {\n")
	imagesToDot(&buffer, roots, byParent)
	buffer.WriteString(" base [style=invisible]\n}\n")

	return buffer.String()
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

func filterImages(images *[]Image, byParent *map[string][]Image) (filteredImages []Image, filteredChildren map[string][]Image) {
	for i := 0; i < len(*images); i++ {
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
			for j := 0; j < len(filteredImages); j++ {
				if filteredImages[j].ParentId == (*images)[i].Id {
					filteredImages[j].ParentId = (*images)[i].ParentId
				}
			}
			for j := 0; j < len(*images); j++ {
				if (*images)[j].ParentId == (*images)[i].Id {
					(*images)[j].ParentId = (*images)[i].ParentId
				}
			}
		}
	}

	filteredChildren = collectChildren(&filteredImages)

	return filteredImages, filteredChildren
}

func jsonToText(buffer *bytes.Buffer, images []Image, byParent map[string][]Image, dispOpts DisplayOpts, prefix string) {
	var length = len(images)
	if length > 1 {
		for index, image := range images {
			var nextPrefix string = ""
			if index+1 == length {
				PrintTreeNode(buffer, image, dispOpts, prefix+"└─")
				nextPrefix = "  "
			} else {
				PrintTreeNode(buffer, image, dispOpts, prefix+"├─")
				nextPrefix = "│ "
			}
			if subimages, exists := byParent[image.Id]; exists {
				jsonToText(buffer, subimages, byParent, dispOpts, prefix+nextPrefix)
			}
		}
	} else {
		for _, image := range images {
			PrintTreeNode(buffer, image, dispOpts, prefix+"└─")
			if subimages, exists := byParent[image.Id]; exists {
				jsonToText(buffer, subimages, byParent, dispOpts, prefix+"  ")
			}
		}
	}
}

func PrintTreeNode(buffer *bytes.Buffer, image Image, dispOpts DisplayOpts, prefix string) {
	var imageID string
	if dispOpts.NoTruncate {
		imageID = image.OrigId
	} else {
		imageID = truncate(stripPrefix(image.OrigId), 12)
	}

	var size int64
	var sizeLabel string
	if dispOpts.Incremental {
		sizeLabel = "Size"
		size = image.Size
	} else {
		sizeLabel = "Virtual Size"
		size = image.VirtualSize
	}

	var sizeStr string
	if dispOpts.NoHuman {
		sizeStr = strconv.FormatInt(size, 10)
	} else {
		sizeStr = humanSize(size)
	}

	buffer.WriteString(fmt.Sprintf("%s%s %s: %s", prefix, imageID, sizeLabel, sizeStr))
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

func truncate(id string, length int) string {
	if len(id) > length {
		return id[0:length]
	} else if len(id) > 0 {
		return id
	} else {
		return ""
	}
}

func stripPrefix(id string) string {
	if strings.Contains(id, ":") {
		idParts := strings.Split(id, ":")
		return idParts[len(idParts)-1]
	}
	return id
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
			buffer.WriteString(fmt.Sprintf(" base -> \"%s\" [style=invis]\n", truncate(image.Id, 12)))
		} else {
			buffer.WriteString(fmt.Sprintf(" \"%s\" -> \"%s\"\n", truncate(image.ParentId, 12), truncate(image.Id, 12)))
		}
		if image.RepoTags[0] != "<none>:<none>" {
			buffer.WriteString(fmt.Sprintf(" \"%s\" [label=\"%s\\n%s\",shape=box,fillcolor=\"paleturquoise\",style=\"filled,rounded\"];\n", truncate(image.Id, 12), truncate(stripPrefix(image.OrigId), 12), strings.Join(image.RepoTags, "\\n")))
		} else {
			// show partial command and size to make up for
			// the fact that since Docker 1.10 content addressing
			// image ids are usually empty and report as <missing>
			SanitizedCommand := SanitizeCommand(image.CreatedBy, 30)
			buffer.WriteString(fmt.Sprintf(" \"%s\" [label=\"%s\"]\n", truncate(image.Id, 12), truncate(stripPrefix(image.OrigId), 12)+"\n"+SanitizedCommand+"\n"+humanize.Bytes(uint64(image.Size))))
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

func SanitizeCommand(CommandStr string, MaxLength int) string {

	temp := CommandStr

	// remove prefixes that don't add meaning
	if strings.HasPrefix(temp, "/bin/sh -c") {
		temp = strings.TrimSpace(temp[10:])
	}
	if strings.HasPrefix(temp, "#(nop)") {
		temp = strings.TrimSpace(temp[6:])
	}

	// remove double and single quotes which make dot format invalid
	temp = strings.Replace(temp, "\"", " ", -1)
	temp = strings.Replace(temp, "'", " ", -1)

	// remove double spaces inside
	temp = strings.Join(strings.Fields(temp), " ")

	return truncate(temp, MaxLength)
}

func init() {
	parser.AddCommand("images",
		"Visualize docker images.",
		"",
		&imagesCommand)
}
