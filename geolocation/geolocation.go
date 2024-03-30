package geolocation

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

var CountriesCodes = []string{
	"ad",
	"ae",
	"af",
	"ag",
	"ai",
	"al",
	"am",
	"ao",
	"ap",
	"aq",
	"ar",
	"as",
	"at",
	"au",
	"aw",
	"ax",
	"az",
	"ba",
	"bb",
	"bd",
	"be",
	"bf",
	"bg",
	"bh",
	"bi",
	"bj",
	"bl",
	"bm",
	"bn",
	"bo",
	"bq",
	"br",
	"bs",
	"bt",
	"bw",
	"by",
	"bz",
	"ca",
	"cd",
	"cf",
	"cg",
	"ch",
	"ci",
	"ck",
	"cl",
	"cm",
	"cn",
	"co",
	"cr",
	"cu",
	"cv",
	"cw",
	"cy",
	"cz",
	"de",
	"dj",
	"dk",
	"dm",
	"do",
	"dz",
	"ec",
	"ee",
	"eg",
	"er",
	"es",
	"et",
	"eu",
	"fi",
	"fj",
	"fk",
	"fm",
	"fo",
	"fr",
	"ga",
	"gb",
	"gd",
	"ge",
	"gf",
	"gg",
	"gh",
	"gi",
	"gl",
	"gm",
	"gn",
	"gp",
	"gq",
	"gr",
	"gt",
	"gu",
	"gw",
	"gy",
	"hk",
	"hn",
	"hr",
	"ht",
	"hu",
	"id",
	"ie",
	"il",
	"im",
	"in",
	"io",
	"iq",
	"ir",
	"is",
	"it",
	"je",
	"jm",
	"jo",
	"jp",
	"ke",
	"kg",
	"kh",
	"ki",
	"km",
	"kn",
	"kp",
	"kr",
	"kw",
	"ky",
	"kz",
	"la",
	"lb",
	"lc",
	"li",
	"lk",
	"lr",
	"ls",
	"lt",
	"lu",
	"lv",
	"ly",
	"ma",
	"mc",
	"md",
	"me",
	"mf",
	"mg",
	"mh",
	"mk",
	"ml",
	"mm",
	"mn",
	"mo",
	"mp",
	"mq",
	"mr",
	"ms",
	"mt",
	"mu",
	"mv",
	"mw",
	"mx",
	"my",
	"mz",
	"na",
	"nc",
	"ne",
	"nf",
	"ng",
	"ni",
	"nl",
	"no",
	"np",
	"nr",
	"nu",
	"nz",
	"om",
	"pa",
	"pe",
	"pf",
	"pg",
	"ph",
	"pk",
	"pl",
	"pm",
	"pr",
	"ps",
	"pt",
	"pw",
	"py",
	"qa",
	"re",
	"ro",
	"rs",
	"ru",
	"rw",
	"sa",
	"sb",
	"sc",
	"sd",
	"se",
	"sg",
	"si",
	"sk",
	"sl",
	"sm",
	"sn",
	"so",
	"sr",
	"ss",
	"st",
	"sv",
	"sx",
	"sy",
	"sz",
	"tc",
	"td",
	"tg",
	"th",
	"tj",
	"tk",
	"tl",
	"tm",
	"tn",
	"to",
	"tr",
	"tt",
	"tv",
	"tw",
	"tz",
	"ua",
	"ug",
	"us",
	"uy",
	"uz",
	"va",
	"vc",
	"ve",
	"vg",
	"vi",
	"vn",
	"vu",
	"wf",
	"ws",
	"ye",
	"yt",
	"za",
	"zm",
	"zw",
	"zz",
}

const filename string = "ipGeolocationCIDR.json"

type IPGeolocation struct {
	CIDRListV4  map[string][]*net.IPNet
	CIDRListV6  map[string][]*net.IPNet
	Ready       bool
	RefreshTime time.Time
}

// Init new `IPGeolocation` and set task to refresh countries_ranges
func New(RefreshPeroidH time.Duration) *IPGeolocation {
	ipGeolocation := IPGeolocation{
		CIDRListV4: make(map[string][]*net.IPNet),
		CIDRListV6: make(map[string][]*net.IPNet),
		Ready:      false,
	}

	go func() {
		err := ipGeolocation.Load(filename)
		if err != nil {
			log.Fatal(err)
		}

		for {
			if !ipGeolocation.Ready || ipGeolocation.RefreshTime.Add(time.Hour*RefreshPeroidH).Before(time.Now()) {
				err := ipGeolocation.RefreshData()
				if err != nil {
					fmt.Println("err in ipGeolocation.RefreshData", err)
				}
				fmt.Println("\nCIDR is READY.")
				time.Sleep(time.Hour * RefreshPeroidH)
			} else {
				fmt.Println("CIDR is Fresh Until", time.Until(ipGeolocation.RefreshTime.Add(time.Hour*RefreshPeroidH)))
				time.Sleep(time.Until(ipGeolocation.RefreshTime.Add(time.Hour * RefreshPeroidH)))
			}
		}
	}()
	return &ipGeolocation
}

// Save serializes IPGeolocation to a JSON file
func (geo *IPGeolocation) Save(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	return json.NewEncoder(file).Encode(geo)
}

// Load deserializes IPGeolocation from a JSON file
func (geo *IPGeolocation) Load(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, geo)
}

// IsDataFresh checks if the data is still fresh (RefreshTime is within the last 60 minutes)
func (geo *IPGeolocation) IsDataFresh() bool {
	return time.Since(geo.RefreshTime).Minutes() < 60
}

func (geo *IPGeolocation) RefreshData() error {
	newCIDRListV4 := make(map[string][]*net.IPNet)
	newCIDRListV6 := make(map[string][]*net.IPNet)

	lenCountriesCodes := len(CountriesCodes)

	for index, code := range CountriesCodes {
		err := geo.downloadCIDRContent("https://raw.githubusercontent.com/onionj/country-ip-blocks-alternative/master/ipv4/"+code+".netset", code, newCIDRListV4)
		if err != nil {
			return err
		}
		err = geo.downloadCIDRContent("https://raw.githubusercontent.com/onionj/country-ip-blocks-alternative/master/ipv6/"+code+".netset", code, newCIDRListV6)
		if err != nil {
			return err
		}
		fmt.Println("Down", code, "CIDR", int32((float32(index+1))/float32(lenCountriesCodes)*100), "% ")
		fmt.Print("\x0D\u001b[1A")
	}

	geo.Ready = true
	geo.RefreshTime = time.Now()
	geo.CIDRListV4 = newCIDRListV4
	geo.CIDRListV6 = newCIDRListV6

	err := geo.Save(filename)
	if err != nil {
		return err
	}

	return nil
}

func (geo *IPGeolocation) downloadCIDRContent(url string, code string, newCIDRList map[string][]*net.IPNet) error {
	response, err := http.Get(url)

	if response.StatusCode != 200 {
		err = fmt.Errorf("failed to download CIDR for country code %s: %d", code, response.StatusCode)
	}
	if err != nil {
		return fmt.Errorf("failed to download CIDR for country code %s: %w", code, err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	scanner := bufio.NewScanner(response.Body)
	for scanner.Scan() {
		_, cidr, _ := net.ParseCIDR(scanner.Text())
		newCIDRList[code] = append(newCIDRList[code], cidr)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read CIDR for country code %s: %w", code, err)
	}

	geo.Ready = true
	geo.RefreshTime = time.Now()
	return nil
}

func (geo *IPGeolocation) Query(ip net.IP) (string, error) {
	if ip.IsPrivate() || ip.IsLoopback() {
		return "", errors.New("IP is private")
	}

	if !geo.Ready {
		return "", errors.New("IPGeolocation is not ready")
	}

	if ip.To4() != nil {
		for country := range geo.CIDRListV4 {
			for _, cidr := range geo.CIDRListV4[country] {
				if cidr.Contains(ip) {
					return country, nil
				}
			}
		}
	} else {
		for country := range geo.CIDRListV6 {
			for _, cidr := range geo.CIDRListV6[country] {
				if cidr.Contains(ip) {
					return country, nil
				}
			}
		}
	}

	return "", errors.New("NotFound")
}
