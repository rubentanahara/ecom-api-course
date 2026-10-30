-- +goose Up
-- +goose StatementBegin
ALTER TABLE order_items ADD CONSTRAINT order_items_quantity_positive CHECK (quantity > 0);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE order_items DROP CONSTRAINT order_items_quantity_positive;
-- +goose StatementEnd
