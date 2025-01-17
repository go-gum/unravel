package examples

import (
	"github.com/go-gum/unravel"
	"github.com/stretchr/testify/require"
	"iter"
	"net/url"
	"testing"
)

type Basket struct {
	PaymentMethod string `json:"pay"`
	ItemIds       []int  `json:"itemId"`
}

func TestUrlValues(t *testing.T) {
	query, _ := url.ParseQuery("itemId=42&itemId=34&itemId=69&pay=Credit")
	t.Logf("Query: %#v", query)

	var basket Basket
	_ = unravel.Unmarshal(UrlValuesSource{Values: query}, &basket)

	require.Equal(t, Basket{ItemIds: []int{42, 34, 69}, PaymentMethod: "Credit"}, basket)
	t.Logf("Mapped Query: %#v", basket)
}

type UrlValuesSource struct {
	unravel.EmptySource
	Values url.Values
}

func (p UrlValuesSource) Get(key string) (unravel.Source, error) {
	return stringSliceSource{Values: p.Values[key]}, nil
}

type stringSliceSource struct {
	unravel.EmptySource
	Values []string
}

func (s stringSliceSource) Int() (int64, error) {
	if len(s.Values) != 1 {
		return 0, unravel.ErrNotSupported
	}
	return unravel.StringSource(s.Values[0]).Int()
}

func (s stringSliceSource) String() (string, error) {
	if len(s.Values) != 1 {
		return "", unravel.ErrNotSupported
	}
	return s.Values[0], nil
}

func (s stringSliceSource) Iter() (iter.Seq[unravel.Source], error) {
	it := func(yield func(unravel.Source) bool) {
		for _, value := range s.Values {
			if !yield(unravel.StringSource(value)) {
				break
			}
		}
	}

	return it, nil
}
