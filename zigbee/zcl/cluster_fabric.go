/*
GSB, 2023
gbatanov@yandex.ru
*/
package zcl

type CommonCluster interface {
	handler_attributes(endpoint Endpoint, attributes []Attribute)
}

func Handler_attributes(cluster CommonCluster, endpoint Endpoint, attributes []Attribute) {
	// cluster.handler_attributes(endpoint, attributes)
}
