-- Migration: Create price_history, price_trends, and price_alerts tables
-- Description: Price tracking system for historical pricing, trend analysis, and user alerts

-- Create price_history table
CREATE TABLE price_history (
    id BIGSERIAL PRIMARY KEY,
    product_master_id INTEGER NOT NULL REFERENCES product_masters(id),
    store_id INTEGER NOT NULL REFERENCES stores(id),
    flyer_id INTEGER REFERENCES flyers(id),

    -- Price information
    price DECIMAL(10,2) NOT NULL,
    original_price DECIMAL(10,2),
    currency VARCHAR(3) DEFAULT 'EUR' NOT NULL,
    is_on_sale BOOLEAN DEFAULT false NOT NULL,

    -- Timing information
    recorded_at TIMESTAMP WITH TIME ZONE NOT NULL,
    valid_from DATE NOT NULL,
    valid_to DATE NOT NULL,
    sale_start_date DATE,
    sale_end_date DATE,

    -- Source information
    source VARCHAR(50) DEFAULT 'flyer' NOT NULL, -- 'flyer', 'manual', 'api'
    extraction_method VARCHAR(50) DEFAULT 'ocr' NOT NULL,
    confidence DECIMAL(3,2) DEFAULT 1.0 NOT NULL,

    -- Availability and stock
    is_available BOOLEAN DEFAULT true NOT NULL,
    stock_level VARCHAR(50),

    -- Metadata
    notes TEXT,
    is_active BOOLEAN DEFAULT true NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,

    -- Constraints
    CONSTRAINT price_history_price_check CHECK (price >= 0),
    CONSTRAINT price_history_original_price_check CHECK (original_price IS NULL OR original_price >= 0),
    CONSTRAINT price_history_confidence_check CHECK (confidence >= 0 AND confidence <= 1),
    CONSTRAINT price_history_valid_dates_check CHECK (valid_to >= valid_from),
    CONSTRAINT price_history_sale_dates_check CHECK (sale_end_date IS NULL OR sale_start_date IS NULL OR sale_end_date >= sale_start_date)
);

-- Create price_trends table for statistical analysis
CREATE TABLE price_trends (
    id BIGSERIAL PRIMARY KEY,
    product_master_id INTEGER NOT NULL REFERENCES product_masters(id),
    store_id INTEGER REFERENCES stores(id),

    -- Period information
    period VARCHAR(10) NOT NULL, -- '7d', '30d', '90d', '1y'
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    calculated_at TIMESTAMP WITH TIME ZONE NOT NULL,

    -- Trend data
    direction VARCHAR(20) NOT NULL, -- 'RISING', 'FALLING', 'STABLE', 'VOLATILE'
    trend_percent DECIMAL(10,2) NOT NULL,
    confidence DECIMAL(3,2) NOT NULL,
    data_points INTEGER NOT NULL,
    volatility_score DECIMAL(5,4) NOT NULL,

    -- Price statistics
    start_price DECIMAL(10,2) NOT NULL,
    end_price DECIMAL(10,2) NOT NULL,
    min_price DECIMAL(10,2) NOT NULL,
    max_price DECIMAL(10,2) NOT NULL,
    avg_price DECIMAL(10,2) NOT NULL,
    median_price DECIMAL(10,2) NOT NULL,

    -- Regression analysis
    slope DECIMAL(10,6),
    intercept DECIMAL(10,2),
    r_squared DECIMAL(3,2),
    is_significant BOOLEAN DEFAULT false NOT NULL,

    -- Moving averages
    ma_7 DECIMAL(10,2),
    ma_14 DECIMAL(10,2),
    ma_30 DECIMAL(10,2),

    -- Metadata
    is_active BOOLEAN DEFAULT true NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,

    -- Constraints
    CONSTRAINT price_trends_confidence_check CHECK (confidence >= 0 AND confidence <= 1),
    CONSTRAINT price_trends_data_points_check CHECK (data_points > 0),
    CONSTRAINT price_trends_volatility_check CHECK (volatility_score >= 0),
    CONSTRAINT price_trends_dates_check CHECK (end_date >= start_date),
    CONSTRAINT price_trends_period_check CHECK (period IN ('7d', '30d', '90d', '1y')),
    CONSTRAINT price_trends_direction_check CHECK (direction IN ('RISING', 'FALLING', 'STABLE', 'VOLATILE'))
);

-- Create price_alerts table for user notifications
CREATE TABLE price_alerts (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_master_id INTEGER NOT NULL REFERENCES product_masters(id),
    store_id INTEGER REFERENCES stores(id),

    -- Alert configuration
    alert_type VARCHAR(20) NOT NULL, -- 'PRICE_DROP', 'TARGET_PRICE', 'PERCENTAGE_DROP'
    target_price DECIMAL(10,2) NOT NULL,
    drop_percent DECIMAL(5,2),
    is_active BOOLEAN DEFAULT true NOT NULL,
    notify_email BOOLEAN DEFAULT true NOT NULL,
    notify_push BOOLEAN DEFAULT false NOT NULL,

    -- Trigger information
    last_triggered TIMESTAMP WITH TIME ZONE,
    trigger_count INTEGER DEFAULT 0 NOT NULL,
    last_price DECIMAL(10,2),

    -- Metadata
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE,

    -- Constraints
    CONSTRAINT price_alerts_target_price_check CHECK (target_price >= 0),
    CONSTRAINT price_alerts_drop_percent_check CHECK (drop_percent IS NULL OR (drop_percent >= 0 AND drop_percent <= 100)),
    CONSTRAINT price_alerts_trigger_count_check CHECK (trigger_count >= 0),
    CONSTRAINT price_alerts_type_check CHECK (alert_type IN ('PRICE_DROP', 'TARGET_PRICE', 'PERCENTAGE_DROP'))
);

-- Create indexes for performance
-- Price History indexes
CREATE INDEX idx_price_history_product_master_id ON price_history(product_master_id);
CREATE INDEX idx_price_history_store_id ON price_history(store_id);
CREATE INDEX idx_price_history_flyer_id ON price_history(flyer_id);
CREATE INDEX idx_price_history_valid_dates ON price_history(valid_from, valid_to);
CREATE INDEX idx_price_history_recorded_at ON price_history(recorded_at);
CREATE INDEX idx_price_history_price ON price_history(price);
CREATE INDEX idx_price_history_is_on_sale ON price_history(is_on_sale) WHERE is_on_sale = true;
CREATE INDEX idx_price_history_active ON price_history(is_active) WHERE is_active = true;

-- Composite index for common query pattern: get price history for product at store
CREATE INDEX idx_price_history_product_store ON price_history(product_master_id, store_id, recorded_at DESC);

-- Index for finding current prices (valid right now)
CREATE INDEX idx_price_history_current ON price_history(product_master_id, store_id, valid_from, valid_to)
    WHERE is_active = true;

-- Price Trends indexes
CREATE INDEX idx_price_trends_product_master_id ON price_trends(product_master_id);
CREATE INDEX idx_price_trends_store_id ON price_trends(store_id);
CREATE INDEX idx_price_trends_period ON price_trends(period);
CREATE INDEX idx_price_trends_calculated_at ON price_trends(calculated_at);
CREATE INDEX idx_price_trends_active ON price_trends(is_active) WHERE is_active = true;

-- Composite index for trend lookup
CREATE INDEX idx_price_trends_product_period ON price_trends(product_master_id, period, calculated_at DESC);

-- Price Alerts indexes
CREATE INDEX idx_price_alerts_user_id ON price_alerts(user_id);
CREATE INDEX idx_price_alerts_product_master_id ON price_alerts(product_master_id);
CREATE INDEX idx_price_alerts_store_id ON price_alerts(store_id);
CREATE INDEX idx_price_alerts_active ON price_alerts(is_active) WHERE is_active = true;
CREATE INDEX idx_price_alerts_expires_at ON price_alerts(expires_at) WHERE expires_at IS NOT NULL;

-- Composite index for user's active alerts
CREATE INDEX idx_price_alerts_user_active ON price_alerts(user_id, is_active) WHERE is_active = true;

-- Function to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_price_trends_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_price_alerts_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for updated_at
CREATE TRIGGER price_trends_updated_at_trigger
BEFORE UPDATE ON price_trends
FOR EACH ROW EXECUTE FUNCTION update_price_trends_updated_at();

CREATE TRIGGER price_alerts_updated_at_trigger
BEFORE UPDATE ON price_alerts
FOR EACH ROW EXECUTE FUNCTION update_price_alerts_updated_at();

-- Add comments for documentation
COMMENT ON TABLE price_history IS 'Historical price data for products from flyers and other sources';
COMMENT ON TABLE price_trends IS 'Calculated price trend statistics and analysis for products';
COMMENT ON TABLE price_alerts IS 'User-configured price alerts for notifications';

COMMENT ON COLUMN price_history.source IS 'Origin of price data: flyer (from scraped flyers), manual (user input), api (external API)';
COMMENT ON COLUMN price_history.extraction_method IS 'Method used to extract price: ocr, manual, api';
COMMENT ON COLUMN price_history.confidence IS 'Confidence score for extracted price (0.0 to 1.0)';

COMMENT ON COLUMN price_trends.direction IS 'Price trend direction: RISING, FALLING, STABLE, or VOLATILE';
COMMENT ON COLUMN price_trends.period IS 'Time period for analysis: 7d, 30d, 90d, 1y';
COMMENT ON COLUMN price_trends.volatility_score IS 'Standard deviation-based volatility measure';

COMMENT ON COLUMN price_alerts.alert_type IS 'Type of alert: PRICE_DROP (any decrease), TARGET_PRICE (reaches target), PERCENTAGE_DROP (% decrease)';
