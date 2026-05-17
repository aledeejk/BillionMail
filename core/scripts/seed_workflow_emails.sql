BEGIN;

INSERT INTO email_templates (temp_name, add_type, content, render, create_time, update_time)
VALUES
(
  'workflow_welcome',
  1,
  '<html><body><h1>Welcome to BillionMail!</h1><p>Hello {{name}}, thank you for subscribing! We are excited to have you.</p></body></html>',
  '',
  EXTRACT(EPOCH FROM NOW())::bigint,
  EXTRACT(EPOCH FROM NOW())::bigint
),
(
  'workflow_reengagement',
  1,
  '<html><body><h1>We miss you!</h1><p>Hi {{name}}, we haven''t seen you for a while. Come back and see what''s new!</p></body></html>',
  '',
  EXTRACT(EPOCH FROM NOW())::bigint,
  EXTRACT(EPOCH FROM NOW())::bigint
),
(
  'workflow_birthday',
  1,
  '<html><body><h1>Happy Birthday, {{name}}!</h1><p>Here''s a special gift just for you. Enjoy your day!</p></body></html>',
  '',
  EXTRACT(EPOCH FROM NOW())::bigint,
  EXTRACT(EPOCH FROM NOW())::bigint
),
(
  'workflow_abandoned_cart',
  1,
  '<html><body><h1>You left something behind!</h1><p>Don''t forget to complete your purchase, {{name}}! Your cart is waiting.</p></body></html>',
  '',
  EXTRACT(EPOCH FROM NOW())::bigint,
  EXTRACT(EPOCH FROM NOW())::bigint
),
(
  'workflow_cross_sale',
  1,
  '<html><body><h1>You might also like this</h1><p>Based on your recent activity, we think you''ll love these picks, {{name}}.</p></body></html>',
  '',
  EXTRACT(EPOCH FROM NOW())::bigint,
  EXTRACT(EPOCH FROM NOW())::bigint
),
(
  'workflow_vip',
  1,
  '<html><body><h1>You''re a VIP, {{name}}!</h1><p>Congratulations — you''ve reached VIP status. Enjoy exclusive perks and early access.</p></body></html>',
  '',
  EXTRACT(EPOCH FROM NOW())::bigint,
  EXTRACT(EPOCH FROM NOW())::bigint
)
ON CONFLICT (temp_name) DO NOTHING;

CREATE TEMP TABLE tpl_ids AS
SELECT id, temp_name FROM email_templates
WHERE temp_name IN (
  'workflow_welcome','workflow_reengagement','workflow_birthday',
  'workflow_abandoned_cart','workflow_cross_sale','workflow_vip'
);

DO $$
DECLARE
  wf              RECORD;
  trigger_node    RECORD;
  old_edge        RECORD;
  tpl_id          INT;
  new_node_id     TEXT;
  new_conn_id     TEXT;
  new_conn2_id    TEXT;
  now_ts          BIGINT := EXTRACT(EPOCH FROM NOW())::BIGINT;
BEGIN

  FOR wf IN SELECT id, name FROM workflow ORDER BY id LOOP

    IF EXISTS (
      SELECT 1 FROM workflow_node
      WHERE workflow_id = wf.id AND type IN ('send-email','email')
    ) THEN
      RAISE NOTICE 'Workflow % (%) already has send-email node – skipped', wf.id, wf.name;
      CONTINUE;
    END IF;

    SELECT id, position_x, position_y
      INTO trigger_node
      FROM workflow_node
     WHERE workflow_id = wf.id AND type = 'trigger'
     ORDER BY created_at ASC LIMIT 1;

    IF trigger_node.id IS NULL THEN
      RAISE NOTICE 'Workflow % (%) has no trigger node – skipped', wf.id, wf.name;
      CONTINUE;
    END IF;

    SELECT id INTO tpl_id FROM tpl_ids
     WHERE LOWER(wf.name) LIKE '%' || REPLACE(REPLACE(temp_name,'workflow_',''),'_',' ') || '%'
     LIMIT 1;

    IF tpl_id IS NULL THEN
      SELECT id INTO tpl_id FROM tpl_ids WHERE
        (LOWER(wf.name) LIKE '%onboard%' OR LOWER(wf.name) LIKE '%welcome%') AND temp_name = 'workflow_welcome'
        LIMIT 1;
    END IF;
    IF tpl_id IS NULL THEN
      SELECT id INTO tpl_id FROM tpl_ids WHERE
        (LOWER(wf.name) LIKE '%re-engage%' OR LOWER(wf.name) LIKE '%reengage%' OR LOWER(wf.name) LIKE '%inactive%') AND temp_name = 'workflow_reengagement'
        LIMIT 1;
    END IF;
    IF tpl_id IS NULL THEN
      SELECT id INTO tpl_id FROM tpl_ids WHERE
        (LOWER(wf.name) LIKE '%birth%') AND temp_name = 'workflow_birthday'
        LIMIT 1;
    END IF;
    IF tpl_id IS NULL THEN
      SELECT id INTO tpl_id FROM tpl_ids WHERE
        (LOWER(wf.name) LIKE '%cart%' OR LOWER(wf.name) LIKE '%abandon%') AND temp_name = 'workflow_abandoned_cart'
        LIMIT 1;
    END IF;
    IF tpl_id IS NULL THEN
      SELECT id INTO tpl_id FROM tpl_ids WHERE
        (LOWER(wf.name) LIKE '%cross%' OR LOWER(wf.name) LIKE '%upsell%') AND temp_name = 'workflow_cross_sale'
        LIMIT 1;
    END IF;
    IF tpl_id IS NULL THEN
      SELECT id INTO tpl_id FROM tpl_ids WHERE
        (LOWER(wf.name) LIKE '%vip%' OR LOWER(wf.name) LIKE '%loyal%') AND temp_name = 'workflow_vip'
        LIMIT 1;
    END IF;
    IF tpl_id IS NULL THEN
      SELECT id INTO tpl_id FROM tpl_ids WHERE temp_name = 'workflow_welcome' LIMIT 1;
    END IF;

    new_node_id  := 'send-email-' || wf.id::TEXT || '-seed';
    new_conn_id  := 'conn-trig-se-' || wf.id::TEXT || '-seed';
    new_conn2_id := 'conn-se-next-' || wf.id::TEXT || '-seed';

    INSERT INTO workflow_node (id, workflow_id, type, config, position_x, position_y, created_at, updated_at)
    VALUES (
      new_node_id,
      wf.id,
      'send-email',
      jsonb_build_object(
        'templateId', tpl_id,
        'sender',     'workflow@billionmail.com',
        'subject',    ''
      ),
      trigger_node.position_x + 250,
      trigger_node.position_y,
      now_ts,
      now_ts
    );

    SELECT id, target INTO old_edge
      FROM workflow_connection
     WHERE workflow_id = wf.id AND source = trigger_node.id
     ORDER BY created_at ASC LIMIT 1;

    INSERT INTO workflow_connection (id, workflow_id, source, target, condition, created_at, updated_at)
    VALUES (new_conn_id, wf.id, trigger_node.id, new_node_id, '', now_ts, now_ts);

    IF old_edge.id IS NOT NULL THEN
      UPDATE workflow_connection
         SET source = new_node_id, updated_at = now_ts
       WHERE id = old_edge.id;
    END IF;

    UPDATE workflow SET version = version + 1, updated_at = now_ts WHERE id = wf.id;

    RAISE NOTICE 'Workflow % (%) → send-email node added (template id=%, template=%)',
      wf.id, wf.name, tpl_id,
      (SELECT temp_name FROM tpl_ids WHERE id = tpl_id);

  END LOOP;
END;
$$;

COMMIT;

SELECT
  w.id          AS workflow_id,
  w.name        AS workflow_name,
  w.version,
  n.id          AS node_id,
  n.type        AS node_type,
  n.config::json->>'templateId' AS template_id,
  t.temp_name   AS template_name
FROM workflow w
JOIN workflow_node n ON n.workflow_id = w.id AND n.type IN ('send-email','email')
LEFT JOIN email_templates t ON t.id = (n.config::json->>'templateId')::INT
ORDER BY w.id;
