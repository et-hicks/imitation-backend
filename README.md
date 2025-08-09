export SUPABASE_DB_URL="postgres://postgres:nJGifZBnAMY4Wq49@https://kaodsutlbtjiffsvmckx.supabase.co/postgres?sslmode=require"

export SUPABASE_DB_URL=postgresql://postgres:nJGifZBnAMY4Wq49@db.kaodsutlbtjiffsvmckx.supabase.co:5432/postgres



export SUPABASE_CLIENT_ANON_KEY=


curl 'https://kaodsutlbtjiffsvmckx.supabase.co/rest/v1/my_table?select=*' \
-H "apikey: SUPABASE_CLIENT_ANON_KEY" \
-H "Authorization: Bearer SUPABASE_CLIENT_ANON_KEY"

		url := os.Getenv("SUPABASE_URL")
		key := os.Getenv("SUPABASE_KEY")

export SUPABASE_URL=https://kaodsutlbtjiffsvmckx.supabase.co
export SUPABASE_KEY=sb_publishable_sXJySn7GqVNHHfvs_A0-aw_LnZRDM2V
