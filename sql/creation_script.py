outfile = "sql/comments.sql"

vals = []
for tid in range(1, 101):
    for k in range(2):
        cid = (tid - 1) * 2 + k + 1
        user_id = ((cid - 1) % 10) + 1
        likes = (cid * 7) % 100
        replies = (cid * 3) % 10
        kind = 'A' if k == 0 else 'B'
        body = f"Comment {kind} on tweet {tid}: thoughts related to tweet #{tid}"
        vals.append("({}, {}, {}, '{}', {}, {})".format(
    cid, user_id, tid, body.replace("'", "''"), likes, replies))

sql = "INSERT INTO comments (id, user_id, tweet_id, body, likes, replies) VALUES\n"
sql += ",\n".join(vals) + ";\n"

with open(outfile, "w", encoding="utf-8") as f:
    f.write(sql)

print(f"Wrote {len(vals)} rows to {outfile}")
