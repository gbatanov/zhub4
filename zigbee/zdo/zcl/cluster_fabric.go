/*
zhub4 - Система домашней автоматизации на Go
Copyright (c) 2023 GSB, Georgii Batanov gbatanov @ yandex.ru
*/

package zcl

type CommonCluster interface {
	Handler_attributes(endpoint Endpoint, attributes []Attribute)
}

func Handler_attributes(cluster CommonCluster, endpoint Endpoint, attributes []Attribute) {
	// cluster.Handler_attributes(endpoint, attributes)
}
