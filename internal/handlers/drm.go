package handlers

import (
	// "bytes"
	"bytes"
	"fmt"
	"time"

	// "regexp"
	// "strconv"
	"strings"
	// "time"

	"github.com/rabilrbl/jiotv_go/v2/pkg/secureurl"
	"github.com/rabilrbl/jiotv_go/v2/pkg/utils"
	"github.com/valyala/fasthttp"

	// "github.com/valyala/fasthttp"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
)

// LiveMpdHandler handles live stream routes /mpd/:channelID
func LiveMpdHandler(c *fiber.Ctx) error {
	// Get channel ID from URL
	channelID := c.Params("channelID")

	// Get live stream URL from JioTV API
	liveResult, err := TV.Live(channelID)
	if err != nil {
		return err
	}
	enc_key, err := secureurl.EncryptURL(liveResult.Mpd.Key)
	if err != nil {
		utils.Log.Panicln(err)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": err,
		})
	}
	
	
	// tv_url, err := secureurl.EncryptURL(liveResult.Mpd.Bitrates.Auto)
	// if err != nil {
	// 	utils.Log.Panicln(err)
	// 	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
	// 		"message": err,
	// 	})
	// }
	tv_url := liveResult.Mpd.Bitrates.Auto
	
	// Send the response
	// return c.JSON(
	// 	fiber.Map{
	// 		"result": liveResult.Mpd,
	// 		"key":	"/drm?auth=" + enc_key + "&channel=" + tv_url + "&channel_id=" + channelID,
	// 	},
	// )
	channel, err := secureurl.EncryptURL(tv_url)
	if err != nil {
		utils.Log.Panicln(err)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": err,
		})
	}
	fmt.Println("deb channel:", tv_url)
	tv_url = strings.Replace(tv_url, "https://jiotvmblive.cdn.jio.com", "", 1)
	return c.Render("views/flow_player_drm", fiber.Map{
		"play_url": tv_url,
		"license_url": "/drm?auth=" + enc_key + "&channel_id=" + channelID + "&channel=" + channel,
	})
}

func generateDateTime() string {
    currentTime := time.Now()
    formattedDateTime := fmt.Sprintf("%02d%02d%02d%02d%02d%03d",
        currentTime.Year()%100, currentTime.Month(), currentTime.Day(),
        currentTime.Hour(), currentTime.Minute(),
        currentTime.Nanosecond()/1000000)
    return formattedDateTime
}


// DRMKeyHandler handles DRM key routes /drm?auth=xxx
func DRMKeyHandler(c *fiber.Ctx) error {
	// Get auth token from URL
	auth := c.Query("auth")
	channel := c.Query("channel")
	channel_id := c.Query("channel_id")

	decoded_channel, err := secureurl.DecryptURL(channel)
	if err != nil {
		utils.Log.Panicln(err)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": err,
		})
	}

	// Make a HEAD request to the decoded_channel to get the cookies
	client := utils.GetRequestClient()
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(decoded_channel)
	req.Header.SetMethod("HEAD")

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Perform the HTTP GET request
	if err := client.Do(req, resp); err != nil {
		utils.Log.Panic(err)
	}

	// Get the cookies from the response
	cookies := resp.Header.Peek("Set-Cookie")
	fmt.Println("Cookie:", string(cookies))

	// Set the cookies in the request
	c.Request().Header.Set("Cookie", string(cookies))


	decoded_url, err := secureurl.DecryptURL(auth)
	if err != nil {
		utils.Log.Panicln(err)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": err,
		})
	}
	fmt.Println(decoded_url)
	// params := strings.Split(decoded_url, "?")[1]
	// // set params as cookies as JioTV uses cookies to authenticate
	// cred_cookie := strings.Split(params, "__hdnea__=")[1]

	// del_regex := `\/bpk-tv\/[^\/]+\/WDVLive`
	// re := regexp.MustCompile(del_regex)

	// cred_cookie = re.ReplaceAllString(cred_cookie, "")

	// exp_regex := `exp=([0-9]+)`
	// re = regexp.MustCompile(exp_regex)
	// match := re.FindStringSubmatch(cred_cookie)
	// expValue := match[1]
	// expUnix, _ := strconv.ParseInt(expValue, 10, 64)
	// expTime := time.Unix(expUnix, 0).UTC()
	// expFormatted := expTime.Format("Mon, 02 Jan 2006 15:04:05 GMT")
	// fmt.Println("Exp value:", expFormatted)

	// fmt.Println(cred_cookie)
	// c.Request().Header.SetCookie("__hdnea__", cred_cookie)
	// c.Request().Header.SetCookie("Domain", "jiotvmblive.cdn.jio.com")
	// c.Request().Header.SetCookie("Path", "/")
	// c.Request().Header.SetCookie("Expires", expFormatted)
	// c.Request().Header.SetCookie("SameSite", "None; Secure")

	
	// Add headers to the request
	c.Request().Header.Set("accesstoken", TV.AccessToken)
	c.Request().Header.Set("Connection", "keep-alive")
	c.Request().Header.Set("os", "android")
	c.Request().Header.Set("appName", "RJIL_JioTV")
	c.Request().Header.Set("subscriberId", TV.Crm)
	c.Request().Header.Set("Host", "tv.media.jio.com")
	c.Request().Header.Set("User-Agent", "plaYtv/7.1.3 (Linux;Android 13) ExoPlayerLib/2.11.7")
	c.Request().Header.Set("ssotoken", TV.SsoToken)
	c.Request().Header.Set("x-platform", "android")
	c.Request().Header.Set("srno", generateDateTime())
	c.Request().Header.Set("crmid", TV.Crm)
	c.Request().Header.Set("channelid", channel_id)
	c.Request().Header.Set("uniqueId", TV.UniqueID)
	c.Request().Header.Set("versionCode", "330")
	c.Request().Header.Set("usergroup", "tvYR7NSNn7rymo3F")
	c.Request().Header.Set("devicetype", "phone")
	c.Request().Header.Set("Accept-Encoding", "gzip, deflate")
	c.Request().Header.Set("osVersion", "13")
	c.Request().Header.Set("deviceId", "b6985e8cf2401d35")
	c.Request().Header.Set("Content-Type", "application/octet-stream")

	// Delete User-Agent header from the request
	c.Request().Header.Del("Accept")
	c.Request().Header.Del("Origin")


	// Print ALL request headers
	utils.Log.Println("Request headers:", c.Request().Header.String())

	if err := proxy.Do(c, decoded_url, TV.Client); err != nil {
		return err
	}

	c.Response().Header.Del(fiber.HeaderServer)
	return nil
}


// BpkProxyHandler handles BPK proxy routes /bpk/:channelID
func BpkProxyHandler(c *fiber.Ctx) error {
	c.Request().Header.Set("Host", "jiotvmblive.cdn.jio.com")
	c.Request().Header.Set("User-Agent", "plaYtv/7.1.3 (Linux;Android 13) ExoPlayerLib/2.11.7")
	
	// Delete headers
	// delete_headers := []string{"Accept", "Sec-Fetch-Dest", "Sec-Fetch-Mode", "Sec-Fetch-Site", "Sec-Fetch-User", "Upgrade-Insecure-Requests", "Accept-Language"}
	
	// for _, header := range delete_headers {
	// 	c.Request().Header.Del(header)
	// }
	
	// Print ALL request headers
	utils.Log.Println("Request headers:", c.Request().Header.String())
	
	// Request path with query params
	url := "https://jiotvmblive.cdn.jio.com" + c.Path() + "?" + string(c.Request().URI().QueryString())
	if url[len(url)-1:] == "?" {
		url = url[:len(url)-1]
	}
	fmt.Println("deb Path:", url)

	if err := proxy.Do(c, url, TV.Client); err != nil {
		return err
	}
	c.Response().Header.Del(fiber.HeaderServer)

	// Delete Domain from cookies
	if c.Response().Header.Peek("Set-Cookie") != nil {
		cookies := c.Response().Header.Peek("Set-Cookie")
		c.Response().Header.Del("Set-Cookie")
		
		cookies = bytes.Replace(cookies, []byte("Domain=jiotvmblive.cdn.jio.com;"), []byte(""), 1)
		// Modify path in cookies
		cookies = bytes.Replace(cookies, []byte("path=/"), []byte("path=/bpk-tv/"), 1)
		
		// Modify Set-Cookie header
		c.Response().Header.SetBytesV("Set-Cookie", cookies)

		fmt.Println("deb Cookies:", string(cookies))
	}

	return nil
}
