desc "Tests all the things"
task :test do
  system "go test ./..."
end

desc "builds the ud and server binaries"
task :build do
  system "rm -rf build/*"

  puts "Building client..."
  system "go build cmd/client/main.go"
  system "mv main build/ud"

  puts "Building server..."
  system "go build cmd/server/main.go"
  system "mv main build/ud-server"
end

task default: :test
