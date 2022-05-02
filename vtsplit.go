package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
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
		if len(startHMS) != 3 {
			fmt.Printf("Each time for song should have 3 parts, got %s\n", s)
			panic("Invalid time part")
		}

		var second float64 = 0

		timepart, err := strconv.ParseFloat(startHMS[0], 64)
		if err != nil {
			panic("Invalid time part")
		}
		second += timepart * 3600

		timepart, err = strconv.ParseFloat(startHMS[1], 64)
		if err != nil {
			panic("Invalid time part")
		}
		second += timepart * 60.0

		timepart, err = strconv.ParseFloat(startHMS[2], 64)
		if err != nil {
			panic("Invalid time part")
		}
		second += timepart

		return second
}

func parse(path string) Info {
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

	var tokenNumber int
	if clipType == "stream" {
		tokenNumber = 4
	} else if clipType == "editted" {
		tokenNumber = 3
	} else {
		panic("Cliptype must be stream or editted")
	}

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", tokenNumber)
		if len(parts) != tokenNumber {
			fmt.Printf("Each line for song should have %d parts, got %s\n", tokenNumber, line)
			continue
		}

		startSecond := parseSecond(parts[1])
		var endSccond = 0.0
		if (clipType == "stream") {
			endSccond = parseSecond(parts[2])
		}

		newSongEnt := SongEnt{
			name:  parts[len(parts) - 1],
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

	return info
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
	var discInfo Info = parse( os.Args[1])
	splitSong(discInfo)
}
