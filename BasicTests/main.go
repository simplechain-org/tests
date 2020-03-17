package main

import (
	"encoding/json"
	"fmt"
	"github.com/simplechain-org/go-simplechain/common"
	"github.com/simplechain-org/go-simplechain/common/math"
	"github.com/simplechain-org/go-simplechain/consensus/scrypt"
	"github.com/simplechain-org/go-simplechain/core/types"
	"github.com/simplechain-org/go-simplechain/params"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
)

type DifficultyTest struct {
	ParentTimestamp    uint64      `json:"parentTimestamp"`
	ParentDifficulty   *big.Int    `json:"parentDifficulty"`
	UncleHash          common.Hash `json:"parentUncles"`
	CurrentTimestamp   uint64      `json:"currentTimestamp"`
	CurrentBlockNumber uint64      `json:"currentBlockNumber"`
	CurrentDifficulty  *big.Int    `json:"currentDifficulty"`
}

type difficultyTestMarshaling struct {
	ParentTimestamp    math.HexOrDecimal64
	ParentDifficulty   *math.HexOrDecimal256
	CurrentTimestamp   math.HexOrDecimal64
	CurrentDifficulty  *math.HexOrDecimal256
	UncleHash          common.Hash
	CurrentBlockNumber math.HexOrDecimal64
}

type Target struct {
	AA map[string]map[string]string
}

func main() {
	path := "/Users/walker/go/src/github.com/simplechain-org/go-simplechain/tests/testdata/BasicTests/difficultyByzantium.json"
	file, err := ReadAll(path)
	if err != nil {
		fmt.Println("error: ", err)
		return
	}

	target := make(map[string]map[string]string)
	if err = json.Unmarshal(file, &target); err != nil {
		fmt.Println("Unmarshal error: ", err)
		return
	}
	//fmt.Println(len(target), target)
	for k, v := range target {
		fmt.Println(k)
		diff := new(DifficultyTest)
		parentTimestamp,err := strconv.ParseUint(strings.TrimLeft(v["parentTimestamp"],"0x"), 16, 64)
		if err !=nil{
			fmt.Println("parentTimestamp error:",err)
		}
		diff.ParentTimestamp =parentTimestamp
		diff.CurrentTimestamp, _ = strconv.ParseUint(strings.TrimLeft(v["currentTimestamp"],"0x"), 16, 64)
		diff.CurrentBlockNumber, _ = strconv.ParseUint(strings.TrimLeft(v["currentBlockNumber"],"0x"), 16, 64)
		ParentDifficulty := new(big.Int)
		ParentDifficulty, _ = ParentDifficulty.SetString(strings.TrimLeft(v["parentDifficulty"], "0x"), 16)
		diff.ParentDifficulty = ParentDifficulty

		CurrentDifficulty := new(big.Int)
		CurrentDifficulty, _ = CurrentDifficulty.SetString(strings.TrimLeft(v["currentDifficulty"], "0x"), 16)
		diff.CurrentDifficulty = CurrentDifficulty
		diff.UncleHash = common.HexToHash(v["parentUncles"])

		if diff.ParentDifficulty.Cmp(params.MinimumDifficulty) < 0 {
			fmt.Println("difficulty below minimum")
			continue
		}
		if err = diff.Run();err!=nil{
			//			fmt.Println(diff.ParentTimestamp,diff.ParentDifficulty.String(),diff.UncleHash.String(),diff.CurrentTimestamp,diff.CurrentBlockNumber)
			fmt.Println(err)
			v["currentDifficulty"] ="0x"+diff.CurrentDifficulty.Text(16)
			//fmt.Println(v["currentDifficulty"])
		}
	}

	//for k, v := range target {
	//	if k == "DifficultyTest2" {
	//		fmt.Println("DifficultyTest2")
	//		v["currentDifficulty"] = "99999"
	//	}
	//}
	data22, _ := json.Marshal(target)
	writeToFile("data1.log", data22)
	fmt.Println("finish")

}

func ReadAll(filePth string) ([]byte, error) {
	f, err := os.Open(filePth)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(f)
}

func readJSONFile(fn string, value interface{}) error {
	file, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer file.Close()

	err = readJSON(file, value)
	if err != nil {
		return fmt.Errorf("%s in file %s", err.Error(), fn)
	}
	return nil
}

func readJSON(reader io.Reader, value interface{}) error {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("error reading JSON file: %v", err)
	}
	if err = json.Unmarshal(data, &value); err != nil {
		if syntaxerr, ok := err.(*json.SyntaxError); ok {
			line := findLine(data, syntaxerr.Offset)
			return fmt.Errorf("JSON syntax error at line %v: %v", line, err)
		}
		return err
	}
	return nil
}

// findLine returns the line number for the given offset into data.
func findLine(data []byte, offset int64) (line int) {
	line = 1
	for i, r := range string(data) {
		if int64(i) >= offset {
			return
		}
		if r == '\n' {
			line++
		}
	}
	return
}

func writeToFile(path string, msg []byte) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Println(err.Error())
	}

	_, err = f.Write(msg)
	if err != nil {
		log.Println(err.Error())
	}
	f.Close()
}

func (test *DifficultyTest) Run() error{
	parentNumber := big.NewInt(int64(test.CurrentBlockNumber - 1))
	parent := &types.Header{
		Difficulty: test.ParentDifficulty,
		Time:       test.ParentTimestamp,
		Number:     parentNumber,
		UncleHash:  test.UncleHash,
	}

	actual := scrypt.CalcDifficulty(nil, test.CurrentTimestamp, parent)
	exp := test.CurrentDifficulty

	if actual.Cmp(exp) != 0 {
		//return fmt.Errorf("parent[time %v diff %v unclehash:%x] child[time %v number %v] diff 0x%x != expected 0x%x",
		//	test.ParentTimestamp, test.ParentDifficulty, test.UncleHash,
		//	test.CurrentTimestamp, test.CurrentBlockNumber, actual, exp)
		//fmt.Printf(" diff 0x%x != expected 0x%x\n", exp, actual)
		test.CurrentDifficulty = actual
		return fmt.Errorf(" diff 0x%x != expected 0x%x", exp, actual)
	}
	return nil

}