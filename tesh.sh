# Check it works
envman --help

# Register
envman register

# Init a project
mkdir testproject && cd testproject
envman init

# Create a test .env
echo "DATABASE_URL=postgres://localhost/test
API_KEY=secret123
PORT=3000" > .env

# Push it
envman push

# Simulate a teammate — delete local .env and get it back
rm .env
envman get

# Check .env was restored
cat .env