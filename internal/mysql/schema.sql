CREATE TABLE IF NOT EXISTS candles(
    id          bigint          auto_increment  primary key,
    symbol      varchar(20)     not null,
    open        double          not null,
    high        double          not null,
    low         double          not null,
    close       double          not null,
    volume      double          not null,
    start_time  datetime        not null,
    end_time    datetime        not null,

    INDEX idx_symbol (symbol),
    INDEX idx_start_time (start_time)
);