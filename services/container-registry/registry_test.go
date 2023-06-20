package containerregistry

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHead(t *testing.T) {
	c := New()
	err := c.Head("997193205088.dkr.ecr.us-east-1.amazonaws.com/source")
	assert.NoError(t, err)
}

func TestGetManifest(t *testing.T) {
	c := New()
	b, err := c.ManifestOrIndex("997193205088.dkr.ecr.us-east-1.amazonaws.com/source")
	assert.NoError(t, err)
	fmt.Printf("string(b): %v\n", string(b))
}

func TestGetIndex(t *testing.T) {
	c := New()
	b, err := c.ManifestOrIndex("cgr.dev/chainguard/nginx")
	assert.NoError(t, err)
	fmt.Printf("string(b): %v\n", string(b))
}

func TestListTag(t *testing.T) {
	c := New()
	tags, err := c.ListTagsWithConstraint("997193205088.dkr.ecr.us-east-1.amazonaws.com/dest", "^1.2.*")
	assert.NoError(t, err)
	fmt.Printf("tags: %v\n", tags)
}

func TestVersionFromSbom(t *testing.T) {
	c := New()
	tag, err := c.VersionFromSbom("nginx", "cgr.dev/chainguard/nginx:1.25.1-r0")
	assert.NoError(t, err)
	fmt.Printf("sboms: %v\n", tag)
}
