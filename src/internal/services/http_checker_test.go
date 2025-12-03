/* Tests pour le service de vérification HTTP
 * Projet de session A25
 * By : Leandre Kanmegne
 * 
 * Teste les différents cas: site dispo, site down, timeout, etc.
 * Pour lancer: go test ./... -race dans la console wsl
 * 
 * Utilise le package net/http/httptest pour créer des serveurs HTTP de test
 * Utilise le package testing de Go pour les assertions et l'organisation des tests
 * Utilise le package context pour gérer les délais d'attente dans les tests
 * Utilise des goroutines pour tester la concurrence
 * 
 * Source: https://pkg.go.dev/net/http/httptest
 */
package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// démarre un faux serveur HTTP généré aleatoirement pour les tests
func creerServeurTest(codeHTTP int, delai time.Duration) *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// simule un site lent
		if delai > 0 {
			time.Sleep(delai)
		}
		w.WriteHeader(codeHTTP)
	})
	return httptest.NewServer(handler)
}

// test : un site qui répond 200 OK
func TestVerifierURL_SiteOK(t *testing.T) {
	serveur := creerServeurTest(http.StatusOK, 0)
	defer serveur.Close()

	ctx, annuler := context.WithTimeout(context.Background(), 5*time.Second)
	defer annuler()

	resultat := VerifierURL(ctx, serveur.URL)

	// doit être marqué comme disponible
	if !resultat.EstDisponible {
		t.Errorf("site devrait être disponible, got EstDisponible=false (code=%d, erreur=%q)",
			resultat.CodeStatutHTTP, resultat.MessageErreur)
	}

	// code HTTP doit être 200
	if resultat.CodeStatutHTTP != http.StatusOK {
		t.Errorf("code HTTP attendu: %d, reçu: %d", http.StatusOK, resultat.CodeStatutHTTP)
	}

	// URL doit correspondre
	if resultat.URL != serveur.URL {
		t.Errorf("URL attendue: %q, reçue: %q", serveur.URL, resultat.URL)
	}
}

// test:avec un site qui retourne 503 Service Unavailable
func TestVerifierURL_SiteDown(t *testing.T) {
	serveur := creerServeurTest(http.StatusServiceUnavailable, 0)
	defer serveur.Close()

	ctx, annuler := context.WithTimeout(context.Background(), 5*time.Second)
	defer annuler()

	resultat := VerifierURL(ctx, serveur.URL)

	// doit être marqué comme indisponible
	if resultat.EstDisponible {
		t.Errorf("site devrait être indisponible avec code 503")
	}

	// code doit être 503
	if resultat.CodeStatutHTTP != http.StatusServiceUnavailable {
		t.Errorf("code attendu: %d, reçu: %d", http.StatusServiceUnavailable, resultat.CodeStatutHTTP)
	}

	// doit avoir un message d'erreur
	if resultat.MessageErreur == "" {
		t.Errorf("MessageErreur devrait pas être vide pour un code 5xx")
	}
}

// test que l'ajout automatique de http:// fonctionne
func TestVerifierURL_AjouteHTTP(t *testing.T) {
	// on teste juste que l'URL est bien préfixée, pas besoin de serveur réel
	ctx, annuler := context.WithTimeout(context.Background(), 1*time.Second)
	defer annuler()

	urlSansPrefix := "example.com"
	resultat := VerifierURL(ctx, urlSansPrefix)

	// doit commencer par http:// ou https://
	prefixeOK := len(resultat.URL) >= 7 && 
		(resultat.URL[:7] == "http://" || 
		(len(resultat.URL) >= 8 && resultat.URL[:8] == "https://"))

	if !prefixeOK {
		t.Errorf("URL devrait commencer par http:// ou https://, got: %q", resultat.URL)
	}
}

// test avec un timeout (site trop lent)
func TestVerifierURL_Timeout(t *testing.T) {
	// serveur qui attend 3 secondes avant de répondre
	serveur := creerServeurTest(http.StatusOK, 3*time.Second)
	defer serveur.Close()

	// timeout de seulement 1 seconde
	ctx, annuler := context.WithTimeout(context.Background(), 1*time.Second)
	defer annuler()

	resultat := VerifierURL(ctx, serveur.URL)

	// doit être marqué comme indisponible à cause du timeout
	if resultat.EstDisponible {
		t.Errorf("site devrait être indisponible à cause du timeout")
	}

	// doit avoir un message d'erreur lié au timeout
	if resultat.MessageErreur == "" {
		t.Errorf("MessageErreur devrait contenir info sur le timeout")
	}
}

// test : plusieurs vérifications en parallèle
func TestVerifierURL_Concurrent(t *testing.T) {
	serveur := creerServeurTest(http.StatusOK, 10*time.Millisecond)
	defer serveur.Close()

	// lance 10 vérifications en parallèle
	nbGoroutines := 10
	termine := make(chan bool, nbGoroutines)

	for i := 0; i < nbGoroutines; i++ {
		go func() {
			ctx, annuler := context.WithTimeout(context.Background(), 5*time.Second)
			defer annuler()

			resultat := VerifierURL(ctx, serveur.URL)
			
			// vérifie que ça marche correctement même en parallèle
			if !resultat.EstDisponible {
				t.Errorf("vérification concurrente échouée")
			}
			
			termine <- true
		}()
	}

	// attend que toutes les goroutines finissent
	for i := 0; i < nbGoroutines; i++ {
		<-termine
	}
}