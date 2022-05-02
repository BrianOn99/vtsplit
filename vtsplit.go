package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"regexp"
)

type SongEnt struct {
	name  string
	start float64
	end   float64
}

type Info struct {
	inputFileName string
	discName      string
	singer        string
	songEnt       []SongEnt
}

func parseSecond(s string) float64 {
		startHMS := strings.Split(s, ":")
		if len(startHMS) < 2 {
			fmt.Printf("Each time for song should have at least 2 parts, got %s\n", s)
			panic("Invalid time part")
		}

		var second float64 = 0

		for len(startHMS) > 0 {
			timepart, err := strconv.ParseFloat(startHMS[0], 64)
			if err != nil {
				panic("Invalid time part")
			}
			second = second*60.0 + timepart
			startHMS = startHMS[1:]
		}

		return second
}

func parse(path string) (Info, error) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
	}

	var info = Info{}
	scanner := bufio.NewScanner(file)
	scanner.Scan()
	info.inputFileName = scanner.Text()
	scanner.Scan()
	info.discName = scanner.Text()
	scanner.Scan()
	info.singer = scanner.Text()
	scanner.Scan()
	var clipType = scanner.Text()
	scanner.Scan()
	var entRegexStr = scanner.Text()
	var re *regexp.Regexp
	re, err = regexp.Compile(entRegexStr)
	if err != nil {
		return info, err
	}

	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindStringSubmatch(line)

		if (matches == nil) {
			return info, fmt.Errorf("Invalid line: %s", line)
		}

		startSecond := parseSecond(matches[re.SubexpIndex("start")])
		var endSccond = 0.0
		if (clipType == "stream") {
			endSccond = parseSecond(matches[re.SubexpIndex("end")])
		}

		newSongEnt := SongEnt{
			name:  matches[re.SubexpIndex("name")],
			start: startSecond,
			end: endSccond,
		}

		if clipType == "editted" {
			if l := len(info.songEnt); l != 0 {
				info.songEnt[l-1].end = startSecond
			}
		}

		info.songEnt = append(info.songEnt, newSongEnt)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}

	cmd := exec.Command(
		"ffprobe", "-i", info.inputFileName,
		"-show_entries", "format=duration",
		"-v", "quiet", "-of", "csv=p=0",
	)

	stdoutStderr, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Println(err)
		panic("error on ffprobe")
	}

	if clipType == "editted" {
		fmt.Sscanf(
			string(stdoutStderr),
			"%f\n",
			&(info.songEnt[len(info.songEnt)-1].end),
		)
	}

	return info, nil
}

func splitSong(discInfo Info) {
	var dirName = strings.Replace(discInfo.discName, "/", "-", -1)
	os.Mkdir(dirName, 0755)

	for index, x := range discInfo.songEnt {
		//ffmpeg -i "a.mp3" -ss 157 -to 572 -c:a copy output.mp3
		fileName := strings.Replace(x.name, "/", "-", -1)
		filePath := fmt.Sprintf("%s/%s.mp3", dirName, fileName)
		println("*****************")
		println(fmt.Sprintf(filePath))
		println("*****************")

		cmd := exec.Command(
			"ffmpeg",
			"-i", discInfo.inputFileName,
			"-ss", strconv.FormatFloat(x.start, 'f', 3, 32),
			"-to", strconv.FormatFloat(x.end, 'f', 3, 32),
			"-c:a", "copy",
			"-metadata", fmt.Sprintf("track=%d", index),
			"-metadata", fmt.Sprintf("album=%s", discInfo.discName),
			"-metadata", fmt.Sprintf("artist=%s", discInfo.singer),
			"-metadata", fmt.Sprintf("title=%s", x.name),
			"-metadata", "genre=vtuber",
			filePath,
		)
		stdoutStderr, err := cmd.CombinedOutput()

		if err != nil {
			fmt.Println(err)
			fmt.Printf("%s\n", stdoutStderr)
		}
	}
}

func main() {
	if (len(os.Args) <2) {
		fmt.Println("No input file")
		os.Exit(-1)
	}
	var discInfo, err = parse( os.Args[1])
	if (err != nil) {
		fmt.Println("Error on parsing file:", err)
		os.Exit(-1)
	}

	splitSong(discInfo)
}
