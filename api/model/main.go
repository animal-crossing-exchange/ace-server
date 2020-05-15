package model

const DatabaseName string = "ace"
const DatabaseURI string = "mongodb://localhost:27017"

func main() {
	// client, err := mongo.NewClient(options.Client().ApplyURI(dbURI))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	// err = client.Connect(ctx)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer client.Disconnect(ctx)
	// err = client.Ping(ctx, readpref.Primary())
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// database := client.Database(dbName)

	// li := NewListingInterface(database)
	// listings, err := li.getListings(ctx)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("%v", listings)
}
