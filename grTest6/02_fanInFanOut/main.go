package main

import (
	"fmt"
	"hash/fnv"
	"sync"
)

const workers = 4

func main() {

	arrayStrings := []string{"AXHxJNwwPwyh", "wBtUOvTUXnEB", "iEIuBrGTcnPY", "ExaKYHVPQJBd", "IisfYYueEgeq", "iCOzJioIdQAW", "TVLFfPwMIMJN", "MlavuLkQKKNA", "dXteBYUCmiIO",
		"LVBYfXKvAPmC", "hFDxTWLHzutJ", "omVNNgFTthdk", "txnkGljmDwAM", "sknPuAIogSTh", "lBhcBoLfxesI", "FoTyOisZKUxJ", "BBuToyAvBNAv", "vfnEnXkeQami", "ONqcMQsXvjeB",
		"kqzYbKUfKPGa", "czbSCvSMVFvS", "LSnDkAxTZSIv", "loyPUmMBGrzY", "gSOnPdCAucii", "kphfyKuMwzgR", "knBfzXdAfuJA", "dAjiFTQRxSHz", "nkzPouILsdgj", "uoibcCuYfZSW", "uukBxcbewfmw",
		"LsuYZkfIwHLd", "BkVDDUUFBKgV", "XGwYQqlUJHfO", "PUeouvpoaJvx", "ISeYHAoTgTzW", "suoJuWELmtJI", "cqYTkAZitXVu", "sqpOAgbGgDCC", "glcSpjhMwnFa", "LpCHviOrtfdB", "lLEjNgRJfwCk",
		"HosgXefTfaCc", "JuOYhmYKWdXY", "XyddKXETFPVc", "JKuTDedipcPL", "amdzINGHJMUd", "zwKJwnMIxhye", "FXGKCxyDJNOZ", "bHcrIQPpFfxv", "HIBIjpPwiSTO", "dbFuKhXGeOAW", "WMENclehPRiq",
		"XNBxsjAgZfJd", "ovEFjSiLDnvu", "FHadEaBUWaQm", "xHbTcbjpUdnz", "rVCgpNsIRmFq", "vvxjuhEcMljO", "kTQwDXGTydQN", "batiuNrHFhBk", "zPhywEfYPYMN", "ktvZSkvlrmYw", "NPOlcYfPiWQy",
		"yXoLgMiGycTn", "KMGZQoCPOrLO", "GaydhaDYPqfs", "PujxdAGuFhdN", "tWZZMPuEPDiI", "DJENOMYostva", "LHnKnnkrtmiS", "yLAkEwuFpMgi", "UpYtJkaqcPIw", "CuZkSQZAHJDI", "uxBcpgcuGtRy",
		"VCYEbBSVoamf", "KVDmNYYpeNkW", "tvgZQcuGWqUx", "SQykIjqqxACg", "gbgCInrgArHC", "QaGYahkxmXlc", "ULeLciRxrDuE", "xClREgpFWDaO", "PmmowLELUHGo", "aOScgZEPPSxQ", "UhzqlSamcjxW",
		"FmVBWNFpPxPF", "yJtlOyuSkcYK", "AAuBLtHrYBCk", "mTpkaRyuWsmt", "PCtdFtkaudBF", "GkhnLkXAFgUd", "LhborCsAvTff", "yLQdwneYeFzw", "FWfekpdDmlNQ", "TNjFXGBeBudF", "XrGTPTeWAfVo",
		"FXrAKnAhszRc", "WuNMObqMSLag", "SooZbGRqTxtu", "rOMTsoekyyfC"}

	var wg sync.WaitGroup
	chIn := make(chan string, workers)
	chOut := make(chan string)
	//в паралель
	chunkSize := (len(arrayStrings) + workers - 1) / workers
	for i := 0; i < workers; i++ {
		startIdx := i * chunkSize
		endIdx := startIdx + chunkSize
		if endIdx > len(arrayStrings) {
			endIdx = len(arrayStrings)
		}
		wg.Add(1)
		go func(chunk []string) {
			defer wg.Done()

			for _, y := range chunk {
				result := fmt.Sprintf("Хеш строки \"%s\": %d", y, stringHash(y))
				chIn <- result
			}
		}(arrayStrings[startIdx:endIdx])
	}

	go func() {
		for data := range chIn {
			chOut <- data
		}
		close(chOut)
	}()

	go func() {
		wg.Wait()
		close(chIn)
	}()

	for out := range chOut {
		fmt.Println(out)
	}

}

func stringHash(s string) uint32 {
	h := fnv.New32a()
	_, err := h.Write([]byte(s))
	if err != nil {
		return 0
	}
	return h.Sum32()
}
