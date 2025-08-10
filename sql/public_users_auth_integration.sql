-- Public users integration with next_auth (keeps your integer PKs)

-- Add public-facing fields if missing
ALTER TABLE public.users
  ADD COLUMN IF NOT EXISTS name  text,
  ADD COLUMN IF NOT EXISTS email text,
  ADD COLUMN IF NOT EXISTS image text;

-- Bridge table: auth UUID -> public integer user id
CREATE TABLE IF NOT EXISTS public.user_auth_map (
  auth_user_id uuid PRIMARY KEY REFERENCES next_auth.users(id) ON DELETE CASCADE,
  user_id      integer UNIQUE NOT NULL REFERENCES public.users(id)
);

-- Trigger to create a public user row and mapping when an auth user is created
CREATE OR REPLACE FUNCTION public.handle_new_auth_user()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
  new_user_id integer;
  uname text;
BEGIN
  uname := COALESCE(NULLIF(split_part(NEW.email, '@', 1), ''), 'user')
           || '-' || substr(NEW.id::text, 1, 8);

  INSERT INTO public.users (username, name, email, image)
  VALUES (uname, NEW.name, NEW.email, NEW.image)
  RETURNING id INTO new_user_id;

  INSERT INTO public.user_auth_map (auth_user_id, user_id)
  VALUES (NEW.id, new_user_id);

  RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS on_auth_user_created ON next_auth.users;
CREATE TRIGGER on_auth_user_created
AFTER INSERT ON next_auth.users
FOR EACH ROW EXECUTE PROCEDURE public.handle_new_auth_user();

-- RLS: Only allow the authenticated user to see/update their own public.users row
ALTER TABLE public.users ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_policies
    WHERE schemaname = 'public' AND tablename = 'users' AND policyname = 'select_own_user'
  ) THEN
    CREATE POLICY select_own_user ON public.users
      FOR SELECT USING (
        EXISTS (
          SELECT 1 FROM public.user_auth_map m
          WHERE m.user_id = users.id AND m.auth_user_id = next_auth.uid()
        )
      );
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_policies
    WHERE schemaname = 'public' AND tablename = 'users' AND policyname = 'update_own_user'
  ) THEN
    CREATE POLICY update_own_user ON public.users
      FOR UPDATE USING (
        EXISTS (
          SELECT 1 FROM public.user_auth_map m
          WHERE m.user_id = users.id AND m.auth_user_id = next_auth.uid()
        )
      );
  END IF;
END$$;


