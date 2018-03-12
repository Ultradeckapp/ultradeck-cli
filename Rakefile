desc "Tests all the things"
task :test do
  system "go test ./..."
end

desc "builds the ud and server binaries"
task :build do
  system "rm -rf build"
  system "mkdir build"

  puts "Building client..."
  system "go build main.go"
  system "mv main build/ultradeck"
  system "cp build/ultradeck /usr/local/bin/"
end

task default: :test
