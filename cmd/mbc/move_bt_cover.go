package main

import (
	"github.com/duke-git/lancet/v2/fileutil"
	"github.com/jeven2016/mylibs/utils"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path"
	"slices"
	"strings"
)

var sourcePicDir *string
var destPicDir *string

var movieFormats = []string{".mp4", ".avi", ".rm"}

func main() {

	var rootCmd = &cobra.Command{
		Version: "0.1",
		Use:     "mbc",
		Short:   "mbc",
		Run: func(cmd *cobra.Command, args []string) {
			movePictures(sourcePicDir, destPicDir)
		},
	}

	sourcePicDir = rootCmd.Flags().StringP("sourceDir", "s", "", "the absolute path of source images")
	destPicDir = rootCmd.Flags().StringP("destDir", "d", "", "the destination path of images")

	if err := rootCmd.Execute(); err != nil {
		utils.PrintCmdErr(err)
	}

}

func movePictures(spd, dpd *string) {
	if spd == nil || dpd == nil {
		log.Printf("source: %v", *sourcePicDir)
		log.Printf("dest: %v", *destPicDir)
		panic("sourcePicDir and destPicDir must be specified")
	}

	// where the pictures should be moved
	sourceFiles, err := os.ReadDir(*spd)
	if err != nil {
		log.Fatal("error occurs", err)
	}

	//where the movies should be there
	destMovieEntries, err := os.ReadDir(*dpd)
	if err != nil {
		log.Fatal("error occurs", err)
	}

	for _, dstFileEntry := range destMovieEntries {
		if dstFileEntry.IsDir() {
			nextDestPath := path.Join(*dpd, dstFileEntry.Name())
			movePictures(spd, &nextDestPath)
			continue
		}
		dstExt := path.Ext(dstFileEntry.Name())
		if !slices.Contains(movieFormats, dstExt) {
			continue
		}

		// dest file name without extension
		dstPureName := strings.ReplaceAll(strings.ReplaceAll(dstFileEntry.Name(), "-", ""), dstExt, "")

		sourceIndex := slices.IndexFunc(sourceFiles, func(entry os.DirEntry) bool {
			if entry.IsDir() {
				return false
			}
			srcPicPureName := strings.ReplaceAll(entry.Name(), path.Ext(entry.Name()), "")
			if strings.Contains(strings.ToLower(dstPureName), strings.ToLower(srcPicPureName)) {
				return true
			}
			return false
		})

		if sourceIndex < 0 {
			continue
		}

		//dest file's path
		destFilePath := path.Join(*dpd, dstPureName+"-poster"+path.Ext(sourceFiles[sourceIndex].Name()))
		if fileutil.IsExist(destFilePath) {
			log.Println("ignored", destFilePath)
			continue
		}

		//source file's path
		sourceFilePath := path.Join(*spd, sourceFiles[sourceIndex].Name())

		err = fileutil.CopyFile(sourceFilePath, destFilePath)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("move", sourceFiles[sourceIndex].Name(), "=>", dstFileEntry.Name(), ":"+destFilePath)
		if err = os.Remove(path.Join(*sourcePicDir, sourceFiles[sourceIndex].Name())); err != nil {
			log.Fatal(err)
		}

	}
}
