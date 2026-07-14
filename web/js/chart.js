let chart;
let candleSeries;

function createChart(){

    const container = document.getElementById("chart");

    chart = LightweightCharts.createChart(
        container,
        {
            width:container.clientWidth,
            height:700,

            layout:{
                background:{
                    color:"#ffffff"
                },
                textColor:"#333333"
            },

            grid:{
                vertLines:{
                    color:"#eef0f2"
                },
                horzLines:{
                    color:"#eef0f2"
                }
            },

            rightPriceScale:{
                borderColor:"#e0e3e7"
            },

            timeScale:{
                borderColor:"#e0e3e7"
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

// const candles = [];

function addCandle(candle){

    candleSeries.update({

        time: Math.floor(
            new Date(candle.StartTime).getTime()/1000
        ),

        open: candle.Open,
        high: candle.High,
        low: candle.Low,
        close: candle.Close,

    });

}

function loadHistory(candles){

    const data = candles.map(c =>({
        time: Math.floor(new Date(c.StartTime).getTime() / 1000),

        open: c.Open,
        high: c.High,
        low: c.Low,
        close: c.Close,
    }));

    candleSeries.setData(data);

}