import { faker } from '@faker-js/faker';
import { writeFile } from 'fs';

interface RequestDTO {
    original_url: string,
    expire_in: number

}

export const Requests: RequestDTO[] = [];

export function createRandomUser(): RequestDTO {
    return {
        original_url: faker.internet.email(),
        expire_in: Math.floor((Math.random() * 24) + 1),
    };
}

Array.from({ length: 1000000 }).forEach(() => {
    Requests.push(createRandomUser());
});
writeFile("data.json", JSON.stringify(Requests), (err) => {
    if (err) throw err;
    console.log('It\'s saved!');
})