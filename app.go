package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
	"github.com/pkg/errors"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func CheckCmdError(cmd *exec.Cmd) error {
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("Error: %v\n", err)
		return err
	}
	if err := cmd.Start(); err != nil {
		log.Printf("Error: %v\n", err)
		return err
	}
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Error: %v\n", err)
	}
	if err := cmd.Wait(); err != nil {
		log.Printf("Error: %v\n", err)
		return err
	}
	return nil
}

// validate
func (a *App) CheckFileExists(path string) error {
	fmt.Printf("check path exists: %s\n", path)
	path = strings.TrimSpace(path)
	if strings.Contains(path, "*") {
		matches, err := filepath.Glob(path)
		if err != nil {
			return err
		}
		if len(matches) == 0 {
			return errors.New("未匹配到任何文件")
		}
		return nil
	}
	if !filepath.IsAbs(path) {
		return errors.New("路径必须是绝对路径!")
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return errors.New("路径不存在!")
	}
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return errors.New("路径是目录!")
	}
	return nil
}

func (a *App) CheckRangeFormat(pages string) error {
	fmt.Printf("check range: %s\n", pages)
	pages = strings.TrimSpace(pages)
	parts := strings.Split(pages, ",")
	pos_count, neg_count := 0, 0
	for _, part := range parts {
		pattern := regexp.MustCompile(`^!?(\d+|N)(\-(\d+|N))?$`)
		part = strings.TrimSpace(part)
		if !pattern.MatchString(part) {
			return errors.New("页码格式错误!,示例：1-3,5,6-N")
		}
		if part[0] == '!' {
			neg_count++
		} else {
			pos_count++
		}
	}
	if pos_count > 0 && neg_count > 0 {
		return errors.New("不能同时使用正向选择和反向选择!")
	}
	return nil
}

// Golang method
func (a *App) CompressPDF(inFile string, outFile string) error {
	if _, err := os.Stat(inFile); os.IsNotExist(err) {
		fmt.Println(err)
		return err
	}
	conf := model.NewDefaultConfiguration()
	if outFile == "" {
		parent := filepath.Dir(inFile)
		parts := strings.Split(filepath.Base(inFile), ".")
		filename := strings.Join(parts[:len(parts)-1], ".")
		outFile = filepath.Join(parent, filename+"_压缩.pdf")
	}
	err := api.OptimizeFile(inFile, outFile, conf)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) ScalePDF(inFile string, outFile string, description string, pagesStr string) error {
	if _, err := os.Stat(inFile); os.IsNotExist(err) {
		fmt.Println(err)
		return err
	}
	pages, err := api.ParsePageSelection(pagesStr)
	if err != nil {
		return err
	}
	resizeConf, err := pdfcpu.ParseResizeConfig(description, types.POINTS)
	if err != nil {
		return err
	}
	conf := model.NewDefaultConfiguration()
	if outFile == "" {
		parent := filepath.Dir(inFile)
		parts := strings.Split(filepath.Base(inFile), ".")
		filename := strings.Join(parts[:len(parts)-1], ".")
		outFile = filepath.Join(parent, filename+"_缩放.pdf")
	}
	err = api.ResizeFile(inFile, outFile, pages, resizeConf, conf)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) ConvertPDF(inFile string, outFile string, dstFormat string, pageStr string) error {
	if _, err := os.Stat(inFile); os.IsNotExist(err) {
		fmt.Println(err)
		return err
	}
	if outFile == "" {
		outFile = filepath.Dir(inFile)
	}
	if _, err := os.Stat(outFile); os.IsNotExist(err) {
		err = os.MkdirAll(outFile, os.ModePerm)
		if err != nil {
			return err
		}
	}
	err := os.Chdir(outFile)
	if err != nil {
		fmt.Println("切换工作目录错误：", err)
		return err
	}
	fmt.Println(outFile)
	path, _ := os.Getwd()
	fmt.Println(path)
	cmd := exec.Command("C:\\Users\\kevin\\code\\wails_demo\\gui_project\\thirdparty\\mutool.exe", "convert", "-F", dstFormat, inFile, pageStr)
	output, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(string(output))

	return nil
}

// Python method
func (a *App) SplitPDFByChunk(inFile string, chunkSize int, outDir string) error {
	fmt.Printf("inFile: %s, chunkSize: %d, outDir: %s\n", inFile, chunkSize, outDir)
	args := []string{"split", "--mode", "chunk"}
	args = append(args, "--chunk_size")
	args = append(args, fmt.Sprintf("%d", chunkSize))
	if outDir != "" {
		args = append(args, "--output", outDir)
	}
	args = append(args, inFile)
	fmt.Printf("%v\n", args)
	fmt.Println(strings.Join(args, ","))
	cmd := exec.Command("C:\\Users\\kevin\\code\\wails_demo\\gui_project\\thirdparty\\dist\\pdf.exe", args...)
	err := CheckCmdError(cmd)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) SplitPDFByBookmark(inFile string, tocLevel string, outDir string) error {
	fmt.Printf("inFile: %s, outDir: %s\n", inFile, outDir)
	args := []string{"split", "--mode", "toc"}
	if tocLevel != "" {
		args = append(args, "--toc-level", tocLevel)
	}
	if outDir != "" {
		args = append(args, "--output", outDir)
	}
	args = append(args, inFile)
	fmt.Printf("%v\n", args)
	fmt.Println(strings.Join(args, ","))
	cmd := exec.Command("C:\\Users\\kevin\\code\\wails_demo\\gui_project\\thirdparty\\dist\\pdf.exe", args...)
	err := CheckCmdError(cmd)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) SplitPDFByPage(inFile string, pages string, outDir string) error {
	fmt.Printf("inFile: %s, pages: %s, outDir: %s\n", inFile, pages, outDir)
	args := []string{"split", "--mode", "page"}
	if pages != "" {
		args = append(args, "--page_range", pages)
	}
	if outDir != "" {
		args = append(args, "--output", outDir)
	}
	args = append(args, inFile)
	fmt.Printf("%v\n", args)
	fmt.Println(strings.Join(args, ","))
	cmd := exec.Command("C:\\Users\\kevin\\code\\wails_demo\\gui_project\\thirdparty\\dist\\pdf.exe", args...)
	err := CheckCmdError(cmd)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) DeletePDF(inFile string, outFile string, pagesStr string) error {
	fmt.Printf("inFile: %s, outFile: %s, pagesStr: %s\n", inFile, outFile, pagesStr)
	args := []string{"delete"}
	if pagesStr != "" {
		args = append(args, "--page_range", pagesStr)
	}
	if outFile != "" {
		args = append(args, "-o", outFile)
	}
	args = append(args, inFile)
	fmt.Printf("%v\n", args)
	cmd := exec.Command("C:\\Users\\kevin\\code\\wails_demo\\gui_project\\thirdparty\\dist\\pdf.exe", args...)
	err := CheckCmdError(cmd)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) InsertPDF(inFile1 string, inFile2 string, insertPos int, dstPages string, outFile string) error {
	fmt.Printf("inFile1: %s, inFile2: %s, insertPos: %d, dstPages: %s, outFile: %s\n", inFile1, inFile2, insertPos, dstPages, outFile)
	args := []string{"insert"}
	if insertPos != 0 {
		args = append(args, "--insert_pos", fmt.Sprintf("%d", insertPos))
	}
	if dstPages != "" {
		args = append(args, "--page_range", dstPages)
	}
	if outFile != "" {
		args = append(args, "-o", outFile)
	}
	args = append(args, inFile1)
	args = append(args, inFile2)
	fmt.Printf("%v\n", args)
	fmt.Println(strings.Join(args, ","))
	cmd := exec.Command("C:\\Users\\kevin\\code\\wails_demo\\gui_project\\thirdparty\\dist\\pdf.exe", args...)
	err := CheckCmdError(cmd)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) ReplacePDF(inFile1 string, inFile2 string, srcPages string, dstPages string, outFile string) error {
	fmt.Printf("inFile1: %s, inFile2: %s, srcPages: %s, dstPages: %s, outFile: %s\n", inFile1, inFile2, srcPages, dstPages, outFile)
	args := []string{"replace"}
	if srcPages != "" {
		args = append(args, "--src_page_range", srcPages)
	}
	if dstPages != "" {
		args = append(args, "--dst_page_range", dstPages)
	}
	if outFile != "" {
		args = append(args, "-o", outFile)
	}
	args = append(args, inFile1)
	args = append(args, inFile2)
	fmt.Printf("%v\n", args)
	fmt.Println(strings.Join(args, ","))
	cmd := exec.Command("C:\\Users\\kevin\\code\\wails_demo\\gui_project\\thirdparty\\dist\\pdf.exe", args...)
	err := CheckCmdError(cmd)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) RotatePDF(inFile string, outFile string, rotation int, pagesStr string) error {
	fmt.Printf("inFile: %s, outFile: %s, rotation: %d, pagesStr: %s\n", inFile, outFile, rotation, pagesStr)
	args := []string{"rotate"}
	if rotation != 0 {
		args = append(args, "--angle", fmt.Sprintf("%d", rotation))
	}
	if pagesStr != "" {
		args = append(args, "--page_range", pagesStr)
	}
	if outFile != "" {
		args = append(args, "-o", outFile)
	}
	args = append(args, inFile)
	fmt.Printf("%v\n", args)
	fmt.Println(strings.Join(args, ","))
	cmd := exec.Command("C:\\Users\\kevin\\code\\wails_demo\\gui_project\\thirdparty\\dist\\pdf.exe", args...)
	err := CheckCmdError(cmd)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) ReorderPDF(inFile string, outFile string, pagesStr string) error {
	fmt.Printf("inFile: %s, outFile: %s, pagesStr: %s\n", inFile, outFile, pagesStr)
	args := []string{"reorder"}
	if pagesStr != "" {
		args = append(args, "--page_range", pagesStr)
	}
	if outFile != "" {
		args = append(args, "-o", outFile)
	}
	args = append(args, inFile)
	fmt.Printf("%v\n", args)
	fmt.Println(strings.Join(args, ","))
	cmd := exec.Command("C:\\Users\\kevin\\code\\wails_demo\\gui_project\\thirdparty\\dist\\pdf.exe", args...)
	err := CheckCmdError(cmd)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) MergePDF(inFiles []string, outFile string, sortMethod string, sortDirection string) error {
	if len(inFiles) == 0 {
		return errors.New("no input files")
	}
	args := []string{"merge"}
	if sortMethod != "" {
		args = append(args, "--sort_method", sortMethod)
	}
	if sortDirection != "" {
		args = append(args, "--sort_direction", sortDirection)
	}
	if outFile != "" {
		args = append(args, "-o", outFile)
	}
	args = append(args, inFiles...)
	fmt.Printf("%v\n", args)
	fmt.Println(strings.Join(args, ","))
	cmd := exec.Command("C:\\Users\\kevin\\code\\wails_demo\\gui_project\\thirdparty\\dist\\pdf.exe", args...)
	err := CheckCmdError(cmd)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) EncryptPDF(inFile string, outFile string, upw string, opw string, perm []string) error {
	fmt.Printf("inFile: %s, outFile: %s, upw: %s, opw: %s, perm: %v\n", inFile, outFile, upw, opw, perm)
	if _, err := os.Stat(inFile); os.IsNotExist(err) {
		fmt.Println(err)
		return err
	}
	args := []string{"encrypt"}
	if len(perm) > 0 {
		args = append(args, "--perm")
		args = append(args, perm...)
	}
	if upw != "" {
		args = append(args, "--user_password", upw)
	}
	if opw != "" {
		args = append(args, "--owner_password", opw)
	}
	if outFile != "" {
		args = append(args, "-o", outFile)
	}
	args = append(args, inFile)
	fmt.Printf("%v\n", args)
	fmt.Println(strings.Join(args, ","))
	cmd := exec.Command("C:\\Users\\kevin\\code\\wails_demo\\gui_project\\thirdparty\\dist\\pdf.exe", args...)
	err := CheckCmdError(cmd)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) DecryptPDF(inFile string, outFile string, passwd string) error {
	fmt.Printf("inFile: %s, outFile: %s, passwd: %s\n", inFile, outFile, passwd)
	if _, err := os.Stat(inFile); os.IsNotExist(err) {
		fmt.Println(err)
		return err
	}
	args := []string{"decrypt"}
	if passwd != "" {
		args = append(args, "--password", passwd)
	}
	if outFile != "" {
		args = append(args, "-o", outFile)
	}
	args = append(args, inFile)
	cmd := exec.Command("C:\\Users\\kevin\\code\\wails_demo\\gui_project\\thirdparty\\dist\\pdf.exe", args...)
	err := CheckCmdError(cmd)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) ExtractBookmark(inFile string, outFile string, format string) error {
	fmt.Printf("inFile: %s, outFile: %s, format: %s\n", inFile, outFile, format)
	if _, err := os.Stat(inFile); os.IsNotExist(err) {
		fmt.Println(err)
		return err
	}
	args := []string{"bookmark", "extract"}
	if format != "" {
		args = append(args, "--format", format)
	}
	if outFile != "" {
		args = append(args, "-o", outFile)
	}
	args = append(args, inFile)
	cmd := exec.Command("C:\\Users\\kevin\\code\\wails_demo\\gui_project\\thirdparty\\dist\\pdf.exe", args...)
	err := CheckCmdError(cmd)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) WriteBookmarkByFile(inFile string, outFile string, tocFile string, offset int) error {
	fmt.Printf("inFile: %s, outFile: %s, tocFile: %s, offset: %d\n", inFile, outFile, tocFile, offset)
	if _, err := os.Stat(inFile); os.IsNotExist(err) {
		fmt.Println(err)
		return err
	}
	if _, err := os.Stat(tocFile); os.IsNotExist(err) {
		fmt.Println(err)
		return err
	}
	args := []string{"bookmark", "add"}
	if tocFile != "" {
		args = append(args, "--toc", tocFile)
	}
	if offset != 0 {
		args = append(args, "--offset", fmt.Sprintf("%d", offset))
	}
	if outFile != "" {
		args = append(args, "-o", outFile)
	}
	args = append(args, inFile)
	cmd := exec.Command("C:\\Users\\kevin\\code\\wails_demo\\gui_project\\thirdparty\\dist\\pdf.exe", args...)
	err := CheckCmdError(cmd)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) WriteBookmarkByGap(inFile string, outFile string, gap int, format string) error {
	fmt.Printf("inFile: %s, outFile: %s, gap: %d\n", inFile, outFile, gap)
	if _, err := os.Stat(inFile); os.IsNotExist(err) {
		fmt.Println(err)
		return err
	}
	args := []string{"bookmark", "add", "--method", "gap"}
	args = append(args, "--gap", fmt.Sprintf("%d", gap))
	if format != "" {
		args = append(args, "--format", format)
	}
	if outFile != "" {
		args = append(args, "-o", outFile)
	}
	args = append(args, inFile)
	cmd := exec.Command("C:\\Users\\kevin\\code\\wails_demo\\gui_project\\thirdparty\\dist\\pdf.exe", args...)
	err := CheckCmdError(cmd)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) TransformBookmark(inFile string, outFile string, addIndent bool, addOffset int, removeDots bool) error {
	fmt.Printf("inFile: %s, outFile: %s, addIndent: %v, addOffset: %d, removeDots: %v\n", inFile, outFile, addIndent, addOffset, removeDots)
	if _, err := os.Stat(inFile); os.IsNotExist(err) {
		fmt.Println(err)
		return err
	}
	args := []string{"bookmark", "transform"}
	if addIndent {
		args = append(args, "--add_indent")
	}
	if addOffset != 0 {
		args = append(args, "--add_offset", fmt.Sprintf("%d", addOffset))
	}
	if removeDots {
		args = append(args, "--remove_trailing_dots")
	}
	if outFile != "" {
		args = append(args, "-o", outFile)
	}
	args = append(args, "--toc", inFile)
	cmd := exec.Command("C:\\Users\\kevin\\code\\wails_demo\\gui_project\\thirdparty\\dist\\pdf.exe", args...)
	err := CheckCmdError(cmd)
	if err != nil {
		return err
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(string(output))
	return nil
}

func (a *App) WatermarkPDF(inFile string, outFile string, markText string, fontFamily string, fontSize int, fontColor string, angle int, space int, opacity float32, quality int) error {
	fmt.Printf("inFile: %s, outFile: %s, markText: %s, fontFamily: %s, fontSize: %d, fontColor: %s, angle: %d, space: %d, opacity: %f, quality: %d\n", inFile, outFile, markText, fontFamily, fontSize, fontColor, angle, space, opacity, quality)
	if _, err := os.Stat(inFile); os.IsNotExist(err) {
		fmt.Println(err)
		return err
	}
	args := []string{"watermark"}
	if markText != "" {
		args = append(args, "--mark-text", markText)
	}
	if fontFamily != "" {
		args = append(args, "--font-family", fontFamily)
	}
	if fontColor != "" {
		args = append(args, "--color", fontColor)
	}
	args = append(args, "--font-size", fmt.Sprintf("%d", fontSize))
	args = append(args, "--angle", fmt.Sprintf("%d", angle))
	args = append(args, "--space", fmt.Sprintf("%d", space))
	args = append(args, "--opacity", fmt.Sprintf("%f", opacity))
	args = append(args, "--quality", fmt.Sprintf("%d", quality))
	if outFile != "" {
		args = append(args, "-o", outFile)
	}
	args = append(args, inFile)
	fmt.Println(args)
	cmd := exec.Command("C:\\Users\\kevin\\code\\wails_demo\\gui_project\\thirdparty\\dist\\pdf.exe", args...)
	err := CheckCmdError(cmd)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) OCR(inFile string, outFile string, pages string, lang string, doubleColumn bool) error {
	if _, err := os.Stat(inFile); os.IsNotExist(err) {
		fmt.Println(err)
		return err
	}
	args := []string{"C:\\Users\\kevin\\code\\wails_demo\\gui_project\\thirdparty\\ocr.py", "ocr"}
	if lang != "" {
		args = append(args, "--lang", lang)
	}
	if doubleColumn {
		args = append(args, "--use-double-column")
	}
	if pages != "" {
		args = append(args, "--range", pages)
	}
	if outFile != "" {
		args = append(args, "-o", outFile)
	}
	args = append(args, inFile)
	fmt.Println(args)
	cmd := exec.Command("C:\\Users\\kevin\\miniconda3\\envs\\ocr\\python.exe", args...)
	err := CheckCmdError(cmd)
	if err != nil {
		return err
	}
	return nil
}
