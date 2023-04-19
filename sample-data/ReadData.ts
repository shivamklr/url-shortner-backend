import { MongoClient } from 'mongodb';
import { writeFile } from 'fs';

const uri = 'mongodb://localhost:27017';
const client = new MongoClient(uri);

async function readData() {
  try {
    await client.connect();
    const database = client.db('golangAPI');
    const collection = database.collection('urls');
    const data = await collection.find().toArray();
    writeFile('ReadData.json', JSON.stringify(data),(err) => {
        if (err) throw err;
        console.log('It\'s saved!');
    });
  } catch (err) {
    console.error(err);
  } finally {
    await client.close();
  }
}

readData();