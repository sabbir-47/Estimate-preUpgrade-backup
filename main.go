package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/shirou/gopsutil/v3/disk"
)

type resource struct {
	name    string
	dirpath string
}

type resourceList struct {
	resources *[]resource
}

type directory struct {
	dirPath    string
	size       float64
	percentage float64
}

type partition struct {
	Device     string
	MountPoint string
	Fstype     string
}

func dirSize(path string) float64 {
	var size float64
	var dirs []string

	err := filepath.Walk(path, func(dirPath string, info os.FileInfo, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrPermission) {
				dirs = append(dirs, dirPath)
			}
			return filepath.SkipDir
		}
		if !info.IsDir() {
			if info.Name() != "/etc/kubernetes/static-pod-resources/bin" {
				size += float64(info.Size())
			}

		}
		return err
	})

	if len(dirs) != 0 {
		fmt.Printf("\nCouldn't fetch disk size for below \ndirectories due to permission denied errors :  \n%s \n", dirs)
	}

	if err != nil {
		size = 0.0
	}

	return size
}

func makeActualDirMap(path string) error {

	actualDirMap := map[string]*directory{}
	var total float64

	dir, err := os.Open(path)
	if err != nil {
		log.Println(err)
		return err
	}
	defer dir.Close()

	files, err := dir.ReadDir(-1)
	if err != nil {
		log.Println(err)
		return err
	}

	for _, file := range files {
		actualDirMap[file.Name()] = &directory{
			dirPath:    fmt.Sprintf("%s/%s", path, file.Name()),
			size:       dirSize(fmt.Sprintf("%s/%s", path, file.Name())),
			percentage: 0.0,
		}
		total += actualDirMap[file.Name()].size
	}

	actualDirMap = calculatePercentage(actualDirMap, total)
	printHeadActual()
	print(actualDirMap, total)

	return nil
}

func sizeConversion(size float64) string {
	units := []string{"B", "KiB", "MiB", "GiB", "TiB", "EiB", "ZiB"}

	i := 0
	if size >= 1024 {
		for i < len(units) && size >= 1024 {
			i++
			size = size / 1024
		}
	}

	return fmt.Sprintf("%.02f %s", size, units[i])
}

func calculatePercentage(m map[string]*directory, total float64) map[string]*directory {
	for key, value := range m {
		value.percentage = value.size * 100 / total
		m[key] = value
	}
	return m
}

func mergeAndDeleteField(m map[string]*directory, path string) map[string]*directory {

	binPathSize := dirSize(path)

	m["staticPods"] = &directory{
		dirPath: m["staticPods"].dirPath,
		size:    m["staticPods"].size - binPathSize,
	}

	m["cluster"] = &directory{
		dirPath: m["cluster"].dirPath,
		size:    m["cluster"].size + m["staticPods"].size,
	}

	delete(m, "staticPods")

	return m
}

func printHeadEstimate() {
	fmt.Printf("\n\n")
	fmt.Println(strings.Repeat("*", 30))
	d := color.New(color.BgBlue)
	d.Printf("%s\n", "Pre-backup estimated disk size\t")
	fmt.Println(strings.Repeat("*", 30))
}

func diskPartitionInfo() {
	fmt.Printf("\n\n")
	fmt.Println(strings.Repeat("*", 19))
	d := color.New(color.BgBlue)
	d.Printf("%s\n", "Disk partition info\t")
	fmt.Println(strings.Repeat("*", 19))
}

func printHeadActual() {
	fmt.Printf("\n\n")
	fmt.Println(strings.Repeat("*", 28))
	d := color.New(color.BgBlue)
	d.Printf("%s\n", "Post-backup actual disk used\t")
	fmt.Println(strings.Repeat("*", 28))
}

func print(m map[string]*directory, total float64) {

	fmt.Println(strings.Repeat("-", 80))
	w := tabwriter.NewWriter(os.Stdout, 10, 0, 0, ' ', tabwriter.Debug)
	fmt.Fprintln(w, "Resource\t Directory\t Size\t Percentage\t")
	for k, v := range m {
		fmt.Fprintln(w, k, "\t", v.dirPath, "\t", sizeConversion(v.size), "\t", fmt.Sprintf("%.2f%%", v.percentage), "\t")
	}
	w.Flush()

	fmt.Println(strings.Repeat("-", 80))
	d := color.New(color.BgBlue)
	d.Printf("%35s", "TOTAL\t")
	d.Printf("= %10s", sizeConversion(total))
	fmt.Printf("\n")
}

func main() {

	//Estimate backup size
	DirList := resourceList{
		resources: &[]resource{
			{"cluster", "/var/lib/etcd/member/snap/db"},
			{"staticPods", "/etc/kubernetes/static-pod-resources/"},
			{"usrLocal", "/usr/local"},
			{"kubelet", "/var/lib/kubelet"},
			{"etc", "/etc"},
		},
	}

	estDirMap := map[string]*directory{}
	var total float64
	for _, v := range *DirList.resources {

		_, err := os.Lstat(v.dirpath)
		if err != nil {
			log.Println(err)
		}

		estDirMap[v.name] = &directory{
			dirPath:    v.dirpath,
			size:       dirSize(v.dirpath),
			percentage: 0.0,
		}
		total += estDirMap[v.name].size

	}

	binPath := "/etc/kubernetes/static-pod-resources/bin/"
	estDirMap = mergeAndDeleteField(estDirMap, binPath)

	estDirMap = calculatePercentage(estDirMap, total)
	printHeadEstimate()
	print(estDirMap, total)

	var testPart []disk.PartitionStat
	var err1 error

	testPart, err1 = disk.Partitions(false)
	if err1 != nil {
		log.Println(err1)
	}

	diskPartitionInfo()
	for _, v := range testPart {
		if v.Mountpoint == "/sysroot" || v.Mountpoint == "/boot" {

			dUsage, err1 := disk.Usage(v.Mountpoint)
			if err1 != nil {
				log.Println(err1)
			}
			fmt.Printf("Device: %s, \t Mountpoint: %s, \t Fstype: %s, \t Total: %s, \t Used: %s, \t UsePercentage: %.2f%%, \t Free: %s \n", v.Device, v.Mountpoint, v.Fstype, sizeConversion(float64(dUsage.Total)), sizeConversion(float64(dUsage.Used)), dUsage.UsedPercent, sizeConversion(float64(dUsage.Free)))
			fmt.Printf("\n\n")
		}
	}

	var recovery string
	if len(os.Args) > 1 {
		if os.Args[1] != "actual" {
			fmt.Printf("Couldn't understand the 2nd parameter, only allowed param is <actual> but found %s\n", os.Args[1])
			os.Exit(1)
		} else {

			recovery = "/var/recovery"
			err := makeActualDirMap(recovery)
			if err != nil {
				log.Println(err)
			}

		}
	}
}
