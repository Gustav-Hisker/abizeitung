function toChart(apidata){
    res = [
        {
            "label":"als beste(r) gewählt",
            "data":[],
            "color":"green"
        },
        {
            "label":"als schlechteste(r) gewählt",
            "data":[],
            "color":"red"
        }
    ]
    for (category in apidata){
        values = apidata[category]
        res[0]["data"].push(values["b"])
        res[1]["data"].push(values["w"])
    }
    return res;
}

exampleOut = [
    {
        "label":"als beste(r) gewählt",
        "data":[14, 35, 45, 45, 45],
        "color":"green"
    },
    {
        "label":"als schlechteste(r) gewählt",
        "data":[5, 46,13, 8, 2],
        "color":"red"
    }
]

exampleData = {"Hausaufgaben":{"b":13,"w":7},"Lustig":{"b":15,"w":4},"Spannender Unterricht":{"b":7,"w":14},"Style":{"b":2,"w":8}}

console.log(toChart(exampleData)) 