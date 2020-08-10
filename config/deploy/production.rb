set :stage, :production
set :disallow_pushing, true
server 's047.srv.iptv2022.com', user: fetch(:user), port: '22', roles: %w(app).freeze
set :keep_releases, 10
