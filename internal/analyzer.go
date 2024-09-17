package internal

import (
	"fmt"
	"path/filepath"
	"strings"
)

func Run(filePath, model string) error {
	if !strings.HasSuffix(filePath, ".tf") {
		return fmt.Errorf("the provided file must have a .tf extension")
	}

	resources, err := ParseTerraformFile(filePath)
	if err != nil {
		return fmt.Errorf("error parsing Terraform file: %v", err)
	}

	dir := filepath.Dir(filePath)
	providerSchema, err := ExtractProviderSchema(dir)
	if err != nil {
		return fmt.Errorf("error extracting provider schema: %v", err)
	}

	return printDiff(resources, providerSchema, model)
}

func printDiff(resources []Resource, schema ProviderSchema, model string) error {
	for _, resource := range resources {
		if possibleAttrs, ok := schema.ResourceTypes[resource.Type]; ok {
			usedAttrs := resource.Attributes
			unusedAttrs := findUnusedAttributes(usedAttrs, possibleAttrs)

			if len(unusedAttrs) > 0 {
				recommendations, err := GetRecommendations(resource.Type, unusedAttrs, model)
				if err != nil {
					return fmt.Errorf("error getting recommendations: %v", err)
				}
				printRecommendations(resource, usedAttrs, recommendations)
			} else {
				fmt.Printf("Resource: %s (Type: %s) - All attributes are used.\n\n", resource.Name, resource.Type)
			}
		}
	}
	return nil
}

func findUnusedAttributes(usedAttrs map[string]string, possibleAttrs map[string]interface{}) []string {
	var unusedAttrs []string
	for attr := range possibleAttrs {
		if _, used := usedAttrs[attr]; !used {
			unusedAttrs = append(unusedAttrs, attr)
		}
	}
	return unusedAttrs
}

func printRecommendations(resource Resource, usedAttrs map[string]string, recommendations string) {
	fmt.Printf("\n# Kuzco-generated recommendations for attributes in %s \"%s\":\n", resource.Type, resource.Name)
	fmt.Printf("resource \"%s\" \"%s\" {\n", resource.Type, resource.Name)

	for attr, value := range usedAttrs {
		fmt.Printf("  %s = \"%s\"\n", attr, value)
	}

	fmt.Printf("\n  # Suggested attributes to enable:\n")
	fmt.Printf("%s", recommendations)
	fmt.Printf("}\n\n")
}
