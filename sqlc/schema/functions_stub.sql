CREATE OR REPLACE FUNCTION public.try_add_delta(p_user_id bigint, p_delta numeric)
RETURNS TABLE (modified boolean, new_amount numeric, old_amount numeric)
LANGUAGE sql
AS $$
    SELECT
        NULL::boolean AS modified,
        NULL::numeric AS new_amount,
        NULL::numeric AS old_amount
$$;