package main

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Pinger struct {
	client *mongo.Client
	opts   *options.ClientOptions
}

func NewPinger(uri string) (*Pinger, error) {
	opts := options.Client().ApplyURI(uri).SetTLSConfig(nil)
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		return nil, err
	}

	return &Pinger{client: client, opts: opts}, nil
}

func (p *Pinger) Close() error {
	return p.client.Disconnect(context.Background())
}

func (p *Pinger) Reconnect() error {
	err := p.client.Disconnect(context.Background())
	if err != nil {
		return err
	}

	client, err := mongo.Connect(context.Background(), p.opts)
	if err != nil {
		return err
	}

	p.client = client
	return nil
}
