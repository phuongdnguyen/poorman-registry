package model

// Image struct
type ImageModel struct {
	// cgr.chainguard.dev/chainguard/nginx:1.25.1-rc.0
	Name string `gorm:"primaryKey"`
	// cgr.chainguard.dev/chainguard/nginx:sha256:81bed54c9e507503766c0f8f030f869705dae486f37c2a003bb5b12bcfcc713f
	HashedIndex string
}
