# Redis schema

## rs1Jokes:{joke_id from rs2}

**type:** string

**description:** A joke.


## rs2JokeIdIndex

**type**: set

**description**: A set containing all the IDs of each joke. Each string in the set represents rs1 > {joke_id}.
