let chart;
let candleSeries;
let volumeSeries;

function createChart(){

    const container = document.getElementById("chart");

    chart = LightweightCharts.createChart(
        container,
        {
            width:container.clientWidth,
            height:container.clientHeight,

            layout:{
                background:{
                    color:"#ffffff"
                },
                textColor:"#4a4f57",
                fontSize:12
            },

            grid:{
                vertLines:{
                    color:"#eef0f2"
                },
                horzLines:{
                    color:"#eef0f2"
                }
            },

            crosshair:{
                mode:LightweightCharts.CrosshairMode.Normal,
                vertLine:{
                    color:"#9598a1",
                    width:1,
                    style:LightweightCharts.LineStyle.Dashed,
                    labelBackgroundColor:"#4a4f57"
                },
                horzLine:{
                    color:"#9598a1",
                    width:1,
                    style:LightweightCharts.LineStyle.Dashed,
                    labelBackgroundColor:"#4a4f57"
                }
            },

            rightPriceScale:{
                borderColor:"#e0e3e7",
                scaleMargins:{
                    top:0.1,
                    bottom:0.25
                }
            },

            timeScale:{
                borderColor:"#e0e3e7",
                timeVisible:true,
                secondsVisible:false
            },

            watermark:{
                visible:true,
                text:"Market Stream",
                color:"rgba(180, 184, 191, 0.35)",
                fontSize:32,
                horzAlign:"center",
                vertAlign:"center"
            }
        }
    );

    candleSeries = chart.addCandlestickSeries({
        upColor:"#2ebd85",
        downColor:"#f6465d",
        borderUpColor:"#2ebd85",
        borderDownColor:"#f6465d",
        wickUpColor:"#2ebd85",
        wickDownColor:"#f6465d"
    });

    volumeSeries = chart.addHistogramSeries({
        priceFormat:{
            type:"volume"
        },
        priceScaleId:"",
        color:"#2ebd85"
    });

    volumeSeries.priceScale().applyOptions({
        scaleMargins:{
            top:0.85,
            bottom:0
        }
    });

    window.addEventListener("resize", () => {
        chart.applyOptions({
            width:container.clientWidth,
            height:container.clientHeight
        });
    });
}

function volumeColor(candle){
    return candle.Close >= candle.Open
        ? "rgba(46, 189, 133, 0.5)"
        : "rgba(246, 70, 93, 0.5)";
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

    const time = Math.floor(
        new Date(candle.StartTime).getTime()/1000
    );

    candleSeries.update({

        time: time,

        open: candle.Open,
        high: candle.High,
        low: candle.Low,
        close: candle.Close,

    });

    volumeSeries.update({
        time: time,
        value: candle.Volume,
        color: volumeColor(candle)
    });

    updateTickerInfo(candle);

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

    volumeSeries.setData(
        candles.map(c => ({
            time: Math.floor(new Date(c.StartTime).getTime() / 1000),
            value: c.Volume,
            color: volumeColor(c)
        }))
    );

    if(candles.length > 0){
        updateTickerInfo(candles[candles.length - 1]);
    }

}

function updateTickerInfo(candle){

    const el = document.getElementById("ticker-ohlc");
    if(!el) return;

    const up = candle.Close >= candle.Open;
    el.classList.toggle("up", up);
    el.classList.toggle("down", !up);

    el.innerHTML =
        `<span>O <b>${candle.Open.toFixed(2)}</b></span>` +
        `<span>H <b>${candle.High.toFixed(2)}</b></span>` +
        `<span>L <b>${candle.Low.toFixed(2)}</b></span>` +
        `<span>C <b>${candle.Close.toFixed(2)}</b></span>` +
        `<span>Vol <b>${candle.Volume.toFixed(3)}</b></span>`;

}