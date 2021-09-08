package connect

import (
	"testing"
)

func TestParseProductsXML(t *testing.T) {
	products, err := parseProductsXML(readTestFile("products.xml", t))
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if len(products) != 2 {
		t.Errorf("Expected len()==2. Got %d", len(products))
	}
	if products[0].ToTriplet() != "SUSE-MicroOS/5.0/x86_64" {
		t.Errorf("Expected SUSE-MicroOS/5.0/x86_64 Got %s", products[0].ToTriplet())
	}
}

func TestParseServicesXML(t *testing.T) {
	services, err := parseServicesXML(readTestFile("services.xml", t))
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if len(services) != 1 {
		t.Errorf("Expected len()==1. Got %d", len(services))
	}
	if services[0].Name != "SUSE_Linux_Enterprise_Micro_5.0_x86_64" {
		t.Errorf("Expected SUSE_Linux_Enterprise_Micro_5.0_x86_64 Got %s", services[0].Name)
	}
}

func TestInstalledProducts(t *testing.T) {
	execute = func(_ []string, _ bool, _ []int) ([]byte, error) {
		return readTestFile("products.xml", t), nil
	}

	products, err := installedProducts()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if len(products) != 2 {
		t.Errorf("Expected len()==2. Got %d", len(products))
	}
	if products[0].ToTriplet() != "SUSE-MicroOS/5.0/x86_64" {
		t.Errorf("Expected SUSE-MicroOS/5.0/x86_64 Got %s", products[0].ToTriplet())
	}
}

func TestBaseProduct(t *testing.T) {
	execute = func(_ []string, _ bool, _ []int) ([]byte, error) {
		return readTestFile("products.xml", t), nil
	}

	base, err := baseProduct()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if base.ToTriplet() != "SUSE-MicroOS/5.0/x86_64" {
		t.Errorf("Expected SUSE-MicroOS/5.0/x86_64 Got %s", base.ToTriplet())
	}
}

func TestBaseProductError(t *testing.T) {
	execute = func(_ []string, _ bool, _ []int) ([]byte, error) {
		return readTestFile("products-no-base.xml", t), nil
	}
	_, err := baseProduct()
	if err != ErrCannotDetectBaseProduct {
		t.Errorf("Unexpected error: %s", err)
	}
}
