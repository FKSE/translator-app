package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"net/http"
	"github.com/labstack/echo/middleware"
)

type Server struct {
	translator *Translator
	echo       *echo.Echo
}

func NewServer(translator *Translator) *Server {
	s := &Server{
		translator: translator,
		echo:       echo.New(),
	}
	// init routing
	s.echo.GET("/languages", s.LanguageList)
	s.echo.POST("/languages", s.LanguageCreate)
	s.echo.GET("/languages/:id", s.LanguageOne)
	s.echo.DELETE("/languages/:id", s.LanguageRemove)
	s.echo.GET("/languages/:lang/translations", s.TranslationList)
	s.echo.POST("/languages/:lang/translations", s.TranslationCreate)
	s.echo.GET("/languages/:lang/translations/:id", s.TranslationOne)

	s.echo.Use(middleware.Logger())

	return s
}

func (s *Server) Run() {
	s.echo.Run(standard.New(":8080"))
}

func (s *Server) LanguageList(c echo.Context) error {
	languages, err := s.translator.Languages()
	if err != nil {
		if err != nil {
			return s.handleError(err)
		}
	}
	return c.JSON(http.StatusOK, languages)
}

func (s *Server) LanguageOne(c echo.Context) error {
	language, err := s.translator.Language(c.Param("id"))
	if err != nil {
		return s.handleError(err)
	}
	return c.JSON(http.StatusOK, language)
}

func (s *Server) LanguageCreate(c echo.Context) error {
	var req map[string]string
	err := c.Bind(&req)
	if err != nil {
		return s.handleError(err)
	}
	lang, err := s.translator.AddLanguage(req["language"], req["baseLanguage"])
	if err != nil {
		return s.handleError(err)
	}
	// save
	err = s.translator.Save(true)
	if err != nil {
		return s.handleError(err)
	}
	return c.JSON(http.StatusOK, lang)
}

func (s *Server) LanguageRemove(c echo.Context) error {
	err := s.translator.RemoveLanguage(c.Param("id"))
	if err != nil {
		return s.handleError(err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (s *Server) TranslationList(c echo.Context) error {
	language, err := s.translator.Language(c.Param("lang"))
	if err != nil {
		return s.handleError(err)
	}
	return c.JSON(http.StatusOK, language["translations"])
}

func (s *Server) TranslationCreate(c echo.Context) error {
	var req map[string]string
	err := c.Bind(&req)
	if err != nil {
		return s.handleError(err)
	}
	err = s.translator.Set(req["id"], req["template"], c.Param("lang"))
	if err != nil {
		return s.handleError(err)
	}
	// save
	err = s.translator.Save(true)
	return c.JSON(http.StatusCreated, req)
}

func (s *Server) TranslationOne(c echo.Context) error {
	language, err := s.translator.Language(c.Param("id"))
	if err != nil {
		return s.handleError(err)
	}
	return c.JSON(http.StatusOK, language)
}

func (s *Server) handleError(err error) *echo.HTTPError {
	if err == ErrLanguageNotFound {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
}