let candleSeries;

function createChart(){

    const chart = LightweightCharts.createChart(
        document.getElementById("chart"),
        {
            width:1200,
            height:700,

            layout:{
                background:{
                    color:"#1e1e1e"
                },
                textColor:"white"
            },

            grid:{
                vertLines:{
                    color:"#333"
                },
                horzLines:{
                    color:"#333"
                }
            }
        }
    );

    candleSeries = chart.addCandlestickSeries();
}

// function addCandle(candle){

//     candleSeries.update({
//         time: Math.floor(
//             new Date(candle.StartTime).getTime()/1000
//         ),

//         open: candle.Open,
//         high: candle.High,
//         low: candle.Low,
//         close: candle.Close,
//     });

// }

const candles = [];

function addCandle(candle){

    candles.push({
        time: Math.floor(new Date(candle.StartTime).getTime()/1000),
        open: candle.Open,
        high: candle.High,
        low: candle.Low,
        close: candle.Close,
    });

    candleSeries.setData(candles);

}