package search

import (
	"context"
	"product-service/models"
	"strings"
	"sync"
)

func SearchByName(ctx context.Context, products []*models.Product, query string) <-chan *models.Product {
	out := make(chan *models.Product, 10)

	go func() {
		defer close(out)
		for _, p := range products {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if strings.Contains(strings.ToLower(p.Name), strings.ToLower(query)) {
				out <- p
			}
		}
	}()
	return out
}

func SearchByCategory(ctx context.Context, products []*models.Product, query string) <-chan *models.Product {
	out := make(chan *models.Product, 10)
	go func() {
		defer close(out)
		for _, p := range products {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if strings.Contains(strings.ToLower(p.Category), strings.ToLower(query)) {
				out <- p
			}
		}
	}()
	return out
}

func SearchByDescription(ctx context.Context, products []*models.Product, query string) <-chan *models.Product {
	out := make(chan *models.Product, 10)
	go func() {
		defer close(out)
		for _, p := range products {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if strings.Contains(strings.ToLower(p.Description), strings.ToLower(query)) {
				out <- p
			}
		}
	}()
	return out
}

func FanIn(ctx context.Context, channels ...<-chan *models.Product) <-chan *models.Product {

	merged := make(chan *models.Product, 30)
	var wg sync.WaitGroup

	forward := func(ch <-chan *models.Product) {
		defer wg.Done()
		for p := range ch {
			select {
			case merged <- p:
			case <-ctx.Done():
				return
			}
		}
	}

	wg.Add(len(channels))

	for _, ch := range channels {
		go forward(ch)
	}

	go func() {
		wg.Wait()
		close(merged)
	}()

	return merged

}
