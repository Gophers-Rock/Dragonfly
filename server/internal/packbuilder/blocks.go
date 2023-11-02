package packbuilder

import (
	"encoding/json"
	"fmt"
	"github.com/df-mc/dragonfly/server/block/customblock"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/segmentio/fasthash/fnv1"
	"image"
	"image/png"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	_ "unsafe" // Imported for compiler directives.
)

var formatVersion = "1.12.0" // TODO: This should be set by the user

// buildBlocks builds all of the block-related files for the resource pack. This includes textures, geometries, language
// entries and terrain texture atlas.
func buildBlocks(dir string) (count int, lang []string) {
	if err := os.MkdirAll(filepath.Join(dir, "models/entity"), os.ModePerm); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "textures/blocks"), os.ModePerm); err != nil {
		panic(err)
	}

	textureData := make(map[string]any)
	for identifier, group := range world.CustomBlocks() {
		if len(group) == 0 {
			panic(fmt.Sprintf("no custom blocks found for identifier %v", identifier))
		}

		base := group[0]
		name := strings.Split(identifier, ":")[1]
		lang = append(lang, fmt.Sprintf("tile.%s.name=%s", identifier, base.Name()))
		textures, permutationTextures, _ := base.Textures()
		for target, texture := range textures {
			textureName := fmt.Sprintf("%s_%s", name, target.Name())
			textureData[textureName] = map[string]string{"textures": "textures/blocks/" + textureName}
			buildBlockTexture(dir, textureName, texture)
		}
		for permutation, permutationSpecificTextures := range permutationTextures {
			h := fnv1.HashString64(permutation)
			for target, texture := range permutationSpecificTextures {
				textureName := fmt.Sprintf("%s_%s_%x", name, target.Name(), h)
				textureData[textureName] = map[string]string{"textures": "textures/blocks/" + textureName}
				buildBlockTexture(dir, textureName, texture)
			}
		}

		buildBlockGeometry(dir, name, base)
		count++
	}

	buildBlockAtlas(dir, map[string]any{
		"resource_pack_name": "vanilla",
		"texture_name":       "atlas.terrain",
		"padding":            8,
		"num_mip_levels":     4,
		"texture_data":       textureData,
	})
	return
}

// buildBlockTexture creates a PNG file for the block from the provided image and name and writes it to the pack.
func buildBlockTexture(dir, name string, img image.Image) {
	texture, err := os.Create(filepath.Join(dir, fmt.Sprintf("textures/blocks/%s.png", name)))
	if err != nil {
		panic(err)
	}
	if err := png.Encode(texture, img); err != nil {
		_ = texture.Close()
		panic(err)
	}
	if err := texture.Close(); err != nil {
		panic(err)
	}
}

// buildBlockGeometry writes the JSON geometry file from the provided name and block and writes it to the pack.
func buildBlockGeometry(dir, name string, block world.CustomBlock) {
	if geometry, permutationGeometries, ok := block.Geometries(); ok {
		data, err := json.Marshal(customblock.Geometries{
			FormatVersion: formatVersion,
			Geometry:      []customblock.Geometry{geometry},
		})
		if err != nil {
			panic(err)
		}
		if err := ioutil.WriteFile(filepath.Join(dir, "models/entity", fmt.Sprintf("%s.geo.json", name)), data, 0666); err != nil {
			panic(err)
		}

		for permutation, permutationSpecificGeometry := range permutationGeometries {
			data, err = json.Marshal(customblock.Geometries{
				FormatVersion: formatVersion,
				Geometry:      []customblock.Geometry{permutationSpecificGeometry},
			})
			if err != nil {
				panic(err)
			}

			h := fnv1.HashString64(permutation)
			if err := ioutil.WriteFile(filepath.Join(dir, "models/entity", fmt.Sprintf("%s_%x.geo.json", name, h)), data, 0666); err != nil {
				panic(err)
			}
		}
	}
}

// buildBlockAtlas creates the identifier to texture mapping and writes it to the pack.
func buildBlockAtlas(dir string, atlas map[string]any) {
	b, err := json.Marshal(atlas)
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(filepath.Join(dir, "textures/terrain_texture.json"), b, 0666); err != nil {
		panic(err)
	}
}
