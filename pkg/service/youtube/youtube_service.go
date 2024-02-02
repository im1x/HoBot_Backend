package youtube

import (
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"io"
	"net/http"
	"regexp"
	"strconv"
)

func GetVideoInfo(id string) (VideoInfo, error) {
	req, err := http.NewRequest("GET", "https://www.youtube.com/watch?v="+id, nil)
	req.Header.Add("sec-ch-ua", " Not A;Brand\";v=\"99\", \"Chromium\";v=\"102\", \"Google Chrome\";v=\"102")
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("upgrade-insecure-requests", "1")
	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.5005.63 Safari/537.36")
	req.Header.Add("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Add("sec-fetch-site", "none")
	req.Header.Add("sec-fetch-mode", "navigate")
	req.Header.Add("sec-fetch-user", "?1")
	req.Header.Add("sec-fetch-dest", "document")
	req.Header.Add("accept-language", "en-US,en;q=0.9")

	q := req.URL.Query()
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("YT request failed:", err)
	}
	defer resp.Body.Close()

	matchJsonRegex, _ := regexp.Compile("var ytInitialPlayerResponse = .+?;</script>")
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("read failed")
	}

	jsonFromBody := matchJsonRegex.FindString(string(b))
	jsonFromBody = jsonFromBody[30 : len(jsonFromBody)-10]

	var videoInfo VideoInfo

	//get title and duration
	re := regexp.MustCompile(regexp.QuoteMeta(`"title":"`) + "(.*?)" + regexp.QuoteMeta(`","lengthSeconds":"`) + "(.*?)" + regexp.QuoteMeta(`"`))
	matches := re.FindStringSubmatch(jsonFromBody)
	videoInfo.Title = matches[1]
	duration, err := strconv.Atoi(matches[2])
	if err != nil {
		log.Error("YT duration parse failed:", err)
		return VideoInfo{}, err
	}
	videoInfo.Duration = duration

	//get views count
	re = regexp.MustCompile(regexp.QuoteMeta(`"viewCount":"`) + "(.*?)" + regexp.QuoteMeta(`"`))
	matches = re.FindStringSubmatch(jsonFromBody)
	viewsCount, err := strconv.Atoi(matches[1])
	if err != nil {
		log.Error("YT viewsCount parse failed:", err)
		return VideoInfo{}, err
	}
	videoInfo.Views = viewsCount

	return videoInfo, nil
}

func ExtractVideoID(url string) (string, error) {
	re := regexp.MustCompile(`(?:youtube\.com/(?:[^/\n\s]+/\S+/|(?:v|e(?:mbed)?)\/|\S*?[?&]v=)|youtu\.be/)([a-zA-Z0-9_-]{11})`)
	match := re.FindStringSubmatch(url)
	if len(match) < 2 {
		return "", fmt.Errorf("unable to extract video ID from the URL")
	}
	fmt.Println(match)
	return match[1], nil
}
