/*
zhub4 - Система домашней автоматизации на Go
Copyright (c) 2023 GSB, Georgii Batanov gbatanov @ yandex.ru
*/

package zcl

type CommonCluster interface {
	HandlerAttributes(endpoint Endpoint, attributes []Attribute)
}

func HandlerAttributes(cluster CommonCluster, endpoint Endpoint, attributes []Attribute) {
	// cluster.HandlerAttributes(endpoint, attributes)
}
