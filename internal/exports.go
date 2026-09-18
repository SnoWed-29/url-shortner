package internal

import (
	"github.com/SnoWed-29/url-shortener/internal/auth"
	"github.com/SnoWed-29/url-shortener/internal/cache"
	"github.com/SnoWed-29/url-shortener/internal/config"
	"github.com/SnoWed-29/url-shortener/internal/database"
	"github.com/SnoWed-29/url-shortener/internal/links"
	"github.com/SnoWed-29/url-shortener/internal/users"
)

type Config = config.Config
type Link = links.Link
type CreateLinkRequest = links.CreateLinkRequest
type User = users.User
type RegisterRequest = users.RegisterRequest
type LoginRequest = users.LoginRequest

var LoadConfig = config.LoadConfig
var NewPostgresPool = database.NewPostgresPool
var NewRedisClient = cache.NewRedisClient
var CacheLink = links.CacheLink
var DeleteCachedLink = links.DeleteCachedLink
var LogCacheError = links.LogCacheError
var GenerateShortCode = links.GenerateShortCode
var ValidateURL = links.ValidateURL
var ValidateCustomAlias = links.ValidateCustomAlias
var CreateLink = links.CreateLink
var GetLinkByShortCode = links.GetLinkByShortCode
var GetLinksByUserID = links.GetLinksByUserID
var CreateUser = users.CreateUser
var GetUserByEmail = users.GetUserByEmail
var HashPassword = auth.HashPassword
var CheckPassword = auth.CheckPassword
var NormalizeEmail = auth.NormalizeEmail
var GenerateJWT = auth.GenerateJWT
var AuthMiddleware = auth.AuthMiddleware
var GetUserID = auth.GetUserID
