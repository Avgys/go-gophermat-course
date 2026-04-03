CREATE OR REPLACE FUNCTION public.try_add_delta(p_user_id bigint, p_delta numeric)
RETURNS TABLE (
    modified boolean,
    new_amount numeric,
    old_amount numeric
)
LANGUAGE plpgsql
AS $$
BEGIN
    -- Ensure a balance row exists
    INSERT INTO public.balance(user_id, amount)
    VALUES (p_user_id, 0)
    ON CONFLICT (user_id) DO NOTHING;

    -- Read current amount
    SELECT b.amount
    INTO old_amount
    FROM public.balance b
    WHERE b.user_id = p_user_id;

    new_amount := old_amount;

    -- Try update
    IF old_amount + p_delta < 0 THEN
        
        
        RAISE EXCEPTION 'insufficient funds: balance % + delta % would be negative', old_amount, p_delta
            USING ERRCODE = '22003';
    END IF;

    UPDATE public.balance b
    SET amount = b.amount + p_delta, withdrawn = CASE WHEN p_delta < 0 THEN b.withdrawn - p_delta ELSE b.withdrawn END
    WHERE b.user_id = p_user_id
    RETURNING b.amount INTO new_amount;

    modified := (new_amount IS NOT NULL);
    RETURN NEXT;
END;
$$;