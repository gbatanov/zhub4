/*
GSB, 2023
gbatanov@yandex.ru
*/
package zcl

type CommonCluster interface {
	Handler_attributes(endpoint Endpoint, attributes []Attribute)
}

func Handler_attributes(cluster CommonCluster, endpoint Endpoint, attributes []Attribute) {
	// cluster.Handler_attributes(endpoint, attributes)
}
