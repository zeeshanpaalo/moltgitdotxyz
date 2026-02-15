UPDATE public.user SET is_admin = true WHERE email = 'ahmadbaigw@gmail.com' RETURNING id, name, email, is_admin;
