package elastic

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/olivere/elastic/v7"
	"strconv"
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

func (es *Es) FuzzyQueryAcc(acc string, acc_length int, acc_type int) (data []string, err error) {
	query := elastic.NewBoolQuery()
	fuzzyEnWordsQuery := elastic.NewFuzzyQuery("acc", acc)
	fuzzyEnWordsQuery.Fuzziness(2)

	query.Must(fuzzyEnWordsQuery)
	notEqAcc := elastic.NewTermsQuery("acc", acc)
	query.MustNot(notEqAcc)
	if acc_length > 0 {
		accLengthTermQuery := elastic.NewTermQuery("acc_length", acc_length)
		query.Must(accLengthTermQuery)
	}
	if acc_type > 0 {
		accTypeTermQuery := elastic.NewTermQuery("acc_type", acc_type)
		query.Must(accTypeTermQuery)
	}

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

	// 读取查询结果
	for _, v := range res.Hits.Hits {
		var temp RecommendAcc
		fmt.Println(v.Id, string(v.Source))
		if err := json.Unmarshal(v.Source, &temp); err != nil {
			continue
		}
		data = append(data, temp.Acc)
	}
	return

}

func (es *Es) TermQueryAcc(word string) (acc RecommendAcc, err error) {
	query := elastic.NewTermQuery("acc", word)
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

	if len(res.Hits.Hits) > 0 {
		if err := json.Unmarshal(res.Hits.Hits[0].Source, &acc); err != nil {
			return acc, err
		}
	}
	return
}
