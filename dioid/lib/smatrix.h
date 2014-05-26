#ifndef __SMATRIX_H__
#define __SMATRIX_H__

class smatrix
{
    private:
        int row, col;
        serie** data;
        void affecte(const serie**, int ligne, int colonne); //affectation d'un ��ent avec un tableau de series

    public:
        smatrix();
        smatrix(int, int); // constructeur 0 : matrice = epsilon + epsilon.(epsilon)*
        smatrix(const smatrix&); // constructuer 1 : matrice initialisee avec un objet matrice
        smatrix(const serie&); // constructuer 2 : matrice initialisee avec un objet serie
        smatrix(poly&); // constructuer 3 : matrice initialisee avec un objet poly
        smatrix(gd&); // constructuer 4 : matrice initialisee avec un objet monome
        ~smatrix(); // destructeur

        int getrow() {return row;}
        int getcol() {return col;}

        smatrix& operator=(const smatrix& a); //surdefinition du =, permet d'initialiser avec une autre matrice
        smatrix& operator=(serie& a); //surdefinition du =, permet d'initialiser avec une serie cast serei matrice
        smatrix& operator=(gd& gd1); // initialise avec un monome cast monome->matrice
        smatrix& operator=(poly& p1); // initialise avec un polynome cast polynome->matrice
        int operator==(const smatrix& M); // surdefiniton de l'�alit�de matrice

        serie& operator ()(int i, int j) {
            if (i >= row || j >= col) {
                taille_incorrecte number(1);
                throw (number);
            }

            return (data[i][j]);
        }

        friend smatrix oplus(smatrix&, smatrix&); //somme de 2 matrices de series p�iodiques
        friend smatrix inf(smatrix& a, smatrix& b); // inf de 2 matrices de s�ies p�iodiques
        friend smatrix otimes(smatrix&, smatrix&); //produit de 2 matrices de series p�iodiques
        friend smatrix lfrac(smatrix&, smatrix&); //residuation a gauche de 2 matrices de series p�iodiques b\a
        friend smatrix rfrac(smatrix& a, smatrix& b); // residuation �droite de 2 matrices de series p�iodiques a/b
        friend smatrix star(smatrix ak_1);

        friend smatrix prcaus(smatrix&);
        friend smatrix odot(smatrix&, smatrix&);
        friend smatrix Duallfrac(smatrix& a, smatrix& b);
};

#endif
