package util

import (
	"encoding/json"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

func GetJson(client *http.Client, url string, target interface{}) error {
	r, err := client.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func FloatToString(f float64) string {
	result := strconv.FormatFloat(f, 'f', 8, 64)

	return result
}

func BigFloatToString(f *big.Float, prec int) string {

	s := f.String()

	r := strings.Split(s, ".")

	switch len(r) {
	case 1:
		return r[0] + ".00"
	case 2:
		if len([]rune(r[1])) == 1 {
			return r[0] + "." + r[1] + "0"
		} else {
			return r[0] + "." + r[1][:prec]
		}
	default:
		return s
	}
}

func DecodeHex(str string) uint64 {
	if len(str) < 2 {
		//log.Errorf("Invalid string: %v", str)
		return 0
	}
	if str == "0x0" || len(str) == 0 {
		return 0
	}

	if str[:2] == "0x" {
		str = str[2:]
	}

	i, err := strconv.ParseUint(str, 16, 64)

	if err != nil {
		log.Errorf("Couldn't decode hex (%v): %v", str, err)
		return 0
	}

	return i
}

func DecodeValueHex(val string) string {
	if val[:2] == "0x" {
		x, err := DecodeBig(val)

		if err != nil {
			log.Errorf("ErrorDecodeValueHex (%v): %v", val, err)
		}
		return x.String()
	} else {
		x, ok := big.NewInt(0).SetString(val, 16)

		if !ok {
			log.Errorf("ErrorDecodeValueHex (%v): %v", val, ok)
		}

		return x.String()
	}
}

func InputParamsToAddress(str string) string {
	return "0x" + strings.ToLower(str[24:])
}

func CaculateBlockReward(height uint64, uncleNo int) *big.Int {

	baseReward := baseBlockReward(height)

	uncleRewards := big.NewInt(0)

	if uncleNo > 0 {
		uncleRewards = uncleRewards.Div(baseReward, big.NewInt(int64(32*uncleNo)))
	}

	baseReward = baseReward.Add(baseReward, uncleRewards)
	return baseReward
}

func CaculateUncleReward(height uint64, uncleHeight uint64) *big.Int {
	baseReward := baseBlockReward(height)

	uncleRewards := big.NewInt(0)

	uncleRewards.Mul(big.NewInt(int64((uncleHeight+2)-height)), baseReward)
	uncleRewards.Div(uncleRewards, big.NewInt(2))

	if uncleRewards.Cmp(big.NewInt(0)) == -1 {
		return big.NewInt(0)
	}
	return uncleRewards
}

func FromWei(str string) string {
	x, _ := new(big.Float).SetString(str)
	y, _ := new(big.Float).SetString("1000000000000000000")
	x.Quo(x, y)
	return x.String()
}

func FromWeiToGwei(str string) string {
	x, _ := new(big.Float).SetString(str)
	y, _ := new(big.Float).SetString("1000000000")
	x.Quo(x, y)
	return x.String()
}

func FormatQwark(str string) string {
	x, _ := new(big.Float).SetString(str)
	y, _ := new(big.Float).SetString("100000000")
	x.Quo(x, y)
	return x.String()
}

func baseBlockReward(height uint64) *big.Int {
	if height > 2508545 {
		return big.NewInt(1000000000000000000)
	} else if height > 2150181 {
		return big.NewInt(2000000000000000000)
	} else if height > 1791818 {
		return big.NewInt(3000000000000000000)
	} else if height > 1433454 {
		return big.NewInt(4000000000000000000)
	} else if height > 1075090 {
		return big.NewInt(5000000000000000000)
	} else if height > 716727 {
		return big.NewInt(6000000000000000000)
	} else if height > 358363 {
		return big.NewInt(7000000000000000000)
	} else if height > 0 {
		return big.NewInt(8000000000000000000)
	} else {
		// genesis
		return big.NewInt(0)
	}
}
