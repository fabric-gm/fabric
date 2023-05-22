package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/pkg/errors"
)

// Detect detects if the provided chaincode source is buildable on Kubernetes
func Detect(ctx context.Context, cfg Config) error {
	log.Println("Procedure: detect")

	if len(os.Args) != 3 {
		return errors.New("detect requires exactly two arguments")
	}

	// Parse metadata.json
	metadataDir := os.Args[2]
	metadata, err := getMetadata(metadataDir)
	if err != nil {
		return errors.Wrap(err, "getting metadata for chaincode")
	}

	// Check if there is a valid image configured
	_, ok := cfg.Images[strings.ToLower(metadata.Type)]
	if !ok {
		return fmt.Errorf("no image available for %q", metadata.Type)
		// Hyperledger Fabric expects a non zero exit code for not
		// detected technologies. main() will ensure a non zero exit code on error
	}

	// Check if platform is supported by hyperledger fabric
	plt := GetPlatform(metadata.Type)
	if plt == nil {
		return fmt.Errorf("platform %q not supported by Hyperledger Fabric", metadata.Type)
	}

	// Image detected successfully
	return nil
}
