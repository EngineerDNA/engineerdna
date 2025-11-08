-- Migration 033: Migrate Cost Configuration to entity_attributes (PDR-9)
-- Consolidates cost_configuration into universal entity_attributes table

-- Migrate cost configuration to entity_attributes
-- Each cost configuration becomes an attribute with temporal validity
INSERT INTO entity_attributes (
    id,
    entity_type,
    entity_id,
    attribute_name,
    value,
    value_type,
    valid_from,
    valid_until,
    source,
    created_at
)
SELECT
    'cost_' || id,
    entity_type,
    COALESCE(entity_id, entity_type),
    'monthly_cost',
    CAST(monthly_cost AS TEXT),
    'currency',
    effective_from,
    effective_to,
    'user_input',
    created_at
FROM cost_configuration
WHERE 'cost_' || id NOT IN (SELECT id FROM entity_attributes);

-- Drop old table (no backward compatibility needed for v1.0)
DROP TABLE IF EXISTS cost_configuration;
