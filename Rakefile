Version = ENV['VERSION']

Envs = [
  {goos: "darwin", arch: "386"},
  {goos: "darwin", arch: "amd64"},
  {goos: "darwin", arch: "arm"},
  {goos: "linux", arch: "386"},
  {goos: "linux", arch: "amd64"},
  {goos: "windows", arch: "386"},
  {goos: "windows", arch: "amd64"}
]

desc "Tests all the things"
task :test do
  system "go test ./..."
end

desc "builds the ud and server binaries"
task :build_local do
  system "rm -rf build"
  system "mkdir build"

  puts "Building client..."
  system "go build main.go"
  system "mv main build/ultradeck"
  system "cp build/ultradeck /usr/local/bin/"
end

desc "Builds release binaries"
task :build_release do
  `rm -rf dist/#{Version}`
  Envs.each do |env|
    ENV["GOOS"] = env[:goos]
    ENV["GOARCH"] = env[:arch]
    puts "Building #{env[:goos]} #{env[:arch]}"
    `GOOS=#{env[:goos]} GOARCH=#{env[:arch]} go build -v -o dist/#{Version}/ultradeck`
    puts "Tarring #{env[:goos]} #{env[:arch]}"
    `tar -czvf dist/#{Version}/ultradeck_#{env[:goos]}_#{env[:arch]}.tar.gz dist/#{Version}/ultradeck`
    puts "Removing dist/#{Version}/ultradeck"
    `rm -rf dist/#{Version}/ultradeck`
  end
end

task default: :test
