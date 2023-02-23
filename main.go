package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/xanzy/go-gitlab"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// TODO rewrite
var gitlabProjectMap = map[string]int{}
var vr = "v0.0.2"
var client *gitlab.Client

type Info struct {
	Version string    `json:"version"` // version string
	Time    time.Time `json:"time"`    // commit time
}

func info(ctx *fiber.Ctx) error {
	if !strings.Contains(ctx.Path(), "youdrive/go/pkg/test-package") {
		return ctx.SendStatus(http.StatusNotFound)
	}
	project := "youdrive/go/pkg/test-package"
	tags, _, err := client.Tags.ListTags(project, &gitlab.ListTagsOptions{})
	if err != nil {
		return ctx.SendStatus(http.StatusNotFound)
	}
	var result []*Info
	for _, tag := range tags {
		result = append(result, &Info{
			Version: tag.Name,
			Time:    *tag.Commit.CommittedDate,
		})
	}

	data, err := json.Marshal(result[0])
	return ctx.Send(data)
}

func archive(ctx *fiber.Ctx) error {
	if !strings.Contains(ctx.Path(), "youdrive/go/pkg/test-package") {
		return ctx.SendStatus(http.StatusNotFound)
	}
	project := "youdrive/go/pkg/test-package"

	format := "zip"
	content, _, err := client.Repositories.Archive(project, &gitlab.ArchiveOptions{
		Format: &format,
		SHA:    &vr,
	})
	if err != nil {
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	reader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	buffer := bytes.NewBuffer([]byte{})

	writer := zip.NewWriter(buffer)

	for _, item := range reader.File {
		parts := strings.Split(item.Name, "/")
		if len(parts) == 0 {
			continue
		}

		directory := "gitlab.citydrive.com6/youdrive/go/pkg/test-package@v0.0.2" //fmt.Sprintf("%s@%s", project, finalVersion)
		file, err := writer.Create(strings.Replace(item.Name, parts[0], directory, 1))
		if err != nil {
			return ctx.SendStatus(http.StatusInternalServerError)
		}

		closer, err := item.Open()
		if err != nil {
			return ctx.SendStatus(http.StatusInternalServerError)
		}
		data, err := ioutil.ReadAll(closer)
		if err != nil {
			return ctx.SendStatus(http.StatusInternalServerError)
		}
		closer.Close()
		if strings.Contains(item.Name, "go.mod") {
			t := string(data)
			t = strings.Replace(t, "module gitlab.citydrive.ru/youdrive/go/pkg/test-package", "module gitlab.citydrive.com6/youdrive/go/pkg/test-package", 1)
			data = []byte(t)
		}
		_, err = file.Write(data)
		if err != nil {
			return ctx.SendStatus(http.StatusInternalServerError)
		}
	}

	err = writer.Close()
	if err != nil {
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	ctx.Response().Header.Set("Content-Length", strconv.FormatInt(int64(buffer.Len()), 10))
	return ctx.SendString(buffer.String())
}

func mod(ctx *fiber.Ctx) error {
	if !strings.Contains(ctx.Path(), "youdrive/go/pkg/test-package") {
		return ctx.SendStatus(http.StatusNotFound)
	}
	project := "youdrive/go/pkg/test-package"

	content, _, err := client.RepositoryFiles.GetRawFile(project, "go.mod", &gitlab.GetRawFileOptions{
		Ref: &vr,
	})
	t := string(content)
	t = strings.Replace(t, "module gitlab.citydrive.ru/youdrive/go/pkg/test-package", "module gitlab.citydrive.com6/youdrive/go/pkg/test-package", 1)
	if err != nil {
		return ctx.SendStatus(http.StatusNotFound)
	}
	return ctx.SendString(t)
}

func list(ctx *fiber.Ctx) error {
	if !strings.Contains(ctx.Path(), "youdrive/go/pkg/test-package") {
		return ctx.SendStatus(http.StatusNotFound)
	}
	project := "youdrive/go/pkg/test-package"
	tags, _, err := client.Tags.ListTags(project, &gitlab.ListTagsOptions{})
	if err != nil {
		return ctx.SendStatus(http.StatusNotFound)
	}
	var result []string
	for _, tag := range tags {
		result = append(result, tag.Name)
	}

	return ctx.SendString(strings.Join(result, "\n"))
}

func init() {
	var err error
	client, err = gitlab.NewClient("glpat-fWkPXj1aW9iRtpAALPKC", gitlab.WithBaseURL("https://gitlab.citydrive.ru"))
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	app := fiber.New()
	app.Get("/*/@v/list", list)
	app.Get("/*/@v/*.mod", mod)
	app.Get("/*/@v/*.zip", archive)
	app.Get("/*/@v/*.info", info)
	app.Listen(":8080")
}
