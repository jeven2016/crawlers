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
var overwrite *bool

var sourceParentDir *string
var destParentDir *string

var movieFormats = []string{".mp4", ".avi", ".rm"}

// 遍历目的文件夹下的子文件夹，在元文件夹下查找匹配的图片并复制到目的文件夹下
func main() {

	var rootCmd = &cobra.Command{
		Version: "0.1",
		Use:     "mbc",
		Short:   "mbc",
		Run: func(cmd *cobra.Command, args []string) {
			log.Println("overwrite:", *overwrite)

			if sourceParentDir != nil && destParentDir != nil {
				moveByParentDir(sourceParentDir, destParentDir)
			} else {
				movePictures(sourcePicDir, destPicDir)
			}
		},
	}

	sourcePicDir = rootCmd.Flags().StringP("sourceDir", "s", "", "the absolute path of source images")
	destPicDir = rootCmd.Flags().StringP("destDir", "d", "", "the destination path of images")
	overwrite = rootCmd.Flags().BoolP("override", "o", false, "overwrite the files")

	sourceParentDir = rootCmd.Flags().StringP("sourceParentDir", "S", "", "the parent absolute path of source images")
	destParentDir = rootCmd.Flags().StringP("destParentDir", "D", "", "the parent destination path of images")

	if err := rootCmd.Execute(); err != nil {
		utils.PrintCmdErr(err)
	}

}

func moveByParentDir(sp *string, dp *string) {
	entries, err := os.ReadDir(*dp)
	if err != nil {
		log.Fatal("failed to read parent destination directory:", err)
	}

	srcEntries, err := os.ReadDir(*sp)
	if err != nil {
		log.Fatal("failed to read parent source directory:", err)
	}

	for _, dstEntry := range entries {
		//check if the name of destination subdirectory exists in sp directory
		if !slices.ContainsFunc(srcEntries, func(e os.DirEntry) bool {
			return e.Name() == dstEntry.Name()
		}) {
			continue
		}

		//both the source and destination directories have the directory in same name
		nextSrcDir := path.Join(*sourceParentDir, dstEntry.Name())
		nextDestDir := path.Join(*destParentDir, dstEntry.Name())
		movePictures(&nextSrcDir, &nextDestDir)
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
		log.Fatal("error occurs:", err)
	}

	//where the movies should be there
	destMovieEntries, err := os.ReadDir(*dpd)
	if err != nil {
		log.Fatal("error occurs:", err)
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
		originDstPicName := strings.ReplaceAll(dstFileEntry.Name(), dstExt, "")

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
		destFilePath := path.Join(*dpd, originDstPicName+"-poster"+path.Ext(sourceFiles[sourceIndex].Name()))
		if !*overwrite && fileutil.IsExist(destFilePath) {
			log.Println("ignored", destFilePath)
			continue
		}

		//correct the wrong file name
		wrongFilePath := path.Join(*dpd, dstPureName+"-poster"+path.Ext(sourceFiles[sourceIndex].Name()))
		if fileutil.IsExist(wrongFilePath) {
			if err = fileutil.CopyFile(wrongFilePath, destFilePath); err != nil {
				log.Fatal("failed to correct file's name", err)
			}
			if err = fileutil.RemoveFile(wrongFilePath); err != nil {
				log.Fatal("failed to remove old source file", err)
			}
			log.Println("correct:", wrongFilePath, "=>", destFilePath)
			continue
		}

		//source file's path
		sourceFilePath := path.Join(*spd, sourceFiles[sourceIndex].Name())

		err = fileutil.CopyFile(sourceFilePath, destFilePath)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("move", sourceFiles[sourceIndex].Name(), "=>", dstFileEntry.Name(), ":"+destFilePath)
		if err = os.Remove(path.Join(*spd, sourceFiles[sourceIndex].Name())); err != nil {
			log.Fatal(err)
		}
		sourceFiles = slices.Delete(sourceFiles, sourceIndex, sourceIndex+1)

		//delete folder picture by default
		folderFile := path.Join(*dpd, "folder.jpg")
		if fileutil.IsExist(folderFile) {
			if err = os.Remove(folderFile); err != nil {
				log.Println("remove folder.jpg", err.Error())
			} else {
				log.Println("removed folder picture:", folderFile)
			}
		}
	}
}
