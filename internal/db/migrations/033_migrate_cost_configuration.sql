-- Migration 033: Migrate Cost Configuration to entity_attributes
-- Consolidates cost_configuration into universal entity_attributes table

-- Migrate cost configuration to entity_attributes
-- Each cost configuration becomes an attribute with temporal validity
-- Role-based costs map to entity_type='role', entity_id=<role_name>
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
    'cost_' || c.id,
    CASE
        WHEN c.role IS NOT NULL THEN 'role'
        ELSE c.entity_type
    END,
    CASE
        WHEN c.role IS NOT NULL THEN c.role
        ELSE COALESCE(c.entity_id, c.entity_type)
    END,
    'monthly_cost',
    CAST(c.monthly_cost AS TEXT),
    'currency',
    c.effective_from,
    c.effective_to,
    'user_input',
    c.created_at
FROM cost_configuration c
WHERE NOT EXISTS (
    SELECT 1 FROM entity_attributes ea
    WHERE ea.entity_type = CASE WHEN c.role IS NOT NULL THEN 'role' ELSE c.entity_type END
    AND ea.entity_id = CASE WHEN c.role IS NOT NULL THEN c.role ELSE COALESCE(c.entity_id, c.entity_type) END
    AND ea.attribute_name = 'monthly_cost'
    AND ea.valid_from = c.effective_from
);

-- Drop old table
DROP TABLE IF EXISTS cost_configuration;
