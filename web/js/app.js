window.onload = async () => {

    createChart();

    await loadHistoryFromServer();

    connectWebSocket();

};