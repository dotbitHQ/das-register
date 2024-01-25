package elastic

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/olivere/elastic/v7"
	"strconv"
	"strings"
)

type RecommendAcc struct {
	AccType  int    `json:"acc_type"`
	AccLenth int    `json:"acc_length"`
	Acc      string `json:"acc"`
}

func (es *Es) InsertRecommendAcc(datas []RecommendAcc) {
	bulkRequest := es.EsCli.Bulk()
	index := "recommend-acc"
	for i, data := range datas {
		doc := elastic.NewBulkIndexRequest().Index(index).Id(strconv.Itoa(i)).Doc(data)
		bulkRequest = bulkRequest.Add(elastic.NewBulkIndexRequest().Index(index).Id(strconv.Itoa(i)).Doc(data))

		bulkRequest = bulkRequest.Add(doc)
	}

	response, err := bulkRequest.Do(context.TODO())
	if err != nil {
		fmt.Println("bulkRequest.Do err : ", err.Error())
		return
	}
	failed := response.Failed()
	iter := len(failed)
	fmt.Printf("error: %v, %v\n", response.Errors, iter)
}

func (es *Es) FuzzyQueryAcc(word string) (data []string, err error) {
	fuzzyEnWordsQuery := elastic.NewFuzzyQuery("acc", word)
	fuzzyEnWordsQuery.Fuzziness(2)
	query := elastic.NewBoolQuery()
	accLengthTermQuery := elastic.NewTermQuery("acc_length", len(word))

	query.Must(fuzzyEnWordsQuery, accLengthTermQuery)
	res, err := es.EsCli.Search().
		Index("recommend-acc").
		Query(query).
		From(0).
		Size(10).
		Do(context.Background())
	if err != nil {
		err = fmt.Errorf("Error performing the search request: %s ", err.Error())
		return
	}
	fmt.Println("Search Results:")
	fmt.Println(strings.Repeat("-", 30))

	// 读取查询结果
	for _, v := range res.Hits.Hits {
		var temp RecommendAcc
		fmt.Println(v.Id, string(v.Source))
		if err := json.Unmarshal(v.Source, &temp); err != nil {
			continue
		}
		fmt.Println(123123, temp)
		data = append(data, temp.Acc)
	}
	return

}
